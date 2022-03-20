package main

import (
	encjson "encoding/json"
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"

	"github.com/spf13/pflag"
	metallbv1beta1 "go.universe.tf/metallb/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/cli-runtime/pkg/resource"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"github.com/sethp/orch/example/02-metallb-kind/manifest"
)

var (
	Args = struct {
		Output string
	}{
		Output: "yaml",
	}

	scheme = runtime.NewScheme()
)

func init() {
	Must := func(fn func(*runtime.Scheme) error) {
		if err := fn(scheme); err != nil {
			panic(err)
		}
	}

	Must(clientgoscheme.AddToScheme)
	Must(metallbv1beta1.AddToScheme)
}

func init() {
	pflag.StringVarP(&Args.Output, "output", "o", Args.Output, "output: yaml or json")
}

func RunCommandLine() error {
	builder := resource.NewLocalBuilder().
		Flatten().
		Unstructured() // TODO

	err := fs.WalkDir(manifest.Files, ".", func(filename string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil
		}

		f, err := manifest.Files.Open(filename)
		if err != nil {
			return fmt.Errorf("reading file %s: %w", filename, err)
		}
		// Doesn't matter, but TODO maybe?
		// defer f.Close()

		builder.Stream(f, filename)

		return nil
	})
	if err != nil {
		return fmt.Errorf("walking manifests: %w", err)
	}

	infos, err := builder.Do().Infos()
	if err != nil {
		return fmt.Errorf("collecting resources: %w", err)
	}

	var objs []runtime.Object
	for _, info := range infos {
		objs = append(objs, info.Object)
	}

	// BEGIN CUSTOM LOGIC
	err = func() error {
		// TODO: reconcile the namespace?
		// TODO: await pod readiness?
		// TODO: crd condition=Established?

		// TODO: `Warning: policy/v1beta1 PodSecurityPolicy is deprecated in v1.21+, unavailable in v1.25+`

		objs = append([]runtime.Object{
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: "metallb-system"},
			},
		}, objs...)

		addrs, err := kindAddressPool()
		if err != nil {
			return err
		}

		// TODO: Error from server (NotFound): error when deleting "STDIN": the server could not find the requested resource (delete addresspools.metallb.io kind)
		objs = append(objs, &metallbv1beta1.AddressPool{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kind",
				Namespace: "metallb-system",
			},
			Spec: metallbv1beta1.AddressPoolSpec{
				Protocol:  "layer2",
				Addresses: addrs,
			},
		})

		return nil
	}()

	if err != nil {
		return fmt.Errorf("doing the thing: %w", err)
	}
	// END CUSTOM LOGIC

	dst := os.Stdout
	var typer runtime.ObjectTyper = scheme

	type Type interface {
		GroupVersionKind() schema.GroupVersionKind
		SetGroupVersionKind(gvk schema.GroupVersionKind)
	}
	for _, obj := range objs {
		o, ok := obj.(Type)
		if !ok {
			return fmt.Errorf("invalid object of type %T, must implement interface Type", obj)
		}

		types, _, err := typer.ObjectKinds(obj)
		switch {
		case runtime.IsNotRegisteredError(err):
			// This'd be snazzier if it suggested `blahv1.AddToScheme`
			return fmt.Errorf("scheme didn't recognize object of type %T (did you forget the relevant group's `AddToScheme`?): %w", obj, err)
		case err != nil:
			return fmt.Errorf("setting type info for %T: %w", obj, err)
		case len(types) != 1:
			return fmt.Errorf("confused by the number of versions (%d, expected 1) for struct %T", len(types), obj)
		}

		o.SetGroupVersionKind(types[0])
	}

	switch Args.Output {
	case "json":
		serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, typer,
			json.SerializerOptions{
				Pretty: true,
			})
		for _, obj := range objs {
			err := serializer.Encode(obj, dst)
			fmt.Fprintln(dst)
			if err != nil {
				return err
			}
		}
		return nil
	case "yaml":
		serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, typer,
			json.SerializerOptions{
				Yaml: true,
			})
		for _, obj := range objs {
			fmt.Fprintln(dst, "---")
			err := serializer.Encode(obj, dst)
			if err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown output format %q", Args.Output)
	}
}

func main() {
	pflag.Parse()
	args := pflag.Args()
	if len(args) < 1 || args[0] != "generate" {
		panic("only generate subcommand implemented")
	}

	if err := RunCommandLine(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func kindAddressPool() ([]string, error) {
	cmd := exec.Command("docker", "network", "inspect", "kind")
	cmd.Stderr = os.Stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer out.Close()

	errCh := make(chan error)
	var networks []struct {
		IPAM struct {
			Config []struct {
				Subnet string
			}
		}
	}
	go func() {
		errCh <- encjson.NewDecoder(out).Decode(&networks)
	}()

	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	if err = <-errCh; err != nil {
		return nil, err
	}

	if len(networks) != 1 {
		return nil, fmt.Errorf("expected exactly one kind network, saw %d", len(networks))
	}

	var addrs []string
	for _, cfg := range networks[0].IPAM.Config {
		_, cidr, err := net.ParseCIDR(cfg.Subnet)
		if err != nil {
			return nil, err
		}

		pool, err := NextSubnet(*cidr)
		if err != nil {
			return nil, err
		}

		addrs = append(addrs, pool.String())
	}

	return addrs, nil
}

// NextSubnet picks a subnet from the middle of the network, to minimize conflicts with
// tools that grow from the bottom up or the top down
//
// Ideally, we'd be able to mark a given IP range as "reserved" for a docker network, but
// alas.
func NextSubnet(cidr net.IPNet) (*net.IPNet, error) {
	ones, bits := cidr.Mask.Size()

	if ones >= bits-1 {
		return nil, fmt.Errorf("not enough addresses")
	}

	nextMask := net.CIDRMask(ones+(bits-ones)/2, bits)
	midpointMask := net.CIDRMask(ones+1, bits)

	n, k := len(cidr.IP), len(cidr.Mask)
	next := make(net.IP, n)
	copy(next, cidr.IP)
	carry := byte(1) // increment the IP by 1
	for i := k - 1; i >= 0; i-- {
		next[(n-k)+i] |= ^midpointMask[i] + carry
		if carry > 0 && next[(n-k)+i] > 0 {
			carry = 0
		}
	}
	return &net.IPNet{
		IP:   next,
		Mask: nextMask,
	}, nil
}
