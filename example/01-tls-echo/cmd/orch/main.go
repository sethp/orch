package main

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"

	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/sethp/orch/example/01-tls-echo/manifest"
)

var (
	Args = struct {
		Output string
	}{
		Output: "yaml",
	}
)

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
		var deploy *unstructured.Unstructured
		for _, obj := range objs {
			o := obj.(*unstructured.Unstructured)
			if o.GetKind() == "Deployment" {
				deploy = o
			}
		}

		if deploy == nil {
			return errors.New("couldn't find deployment")
		}

		certBytes, keyBytes, err := func() (certBytes []byte, keyBytes []byte, err error) {
			// Or, consider: generating a certificate here
			certBytes, err = ioutil.ReadFile("server.crt")
			if err != nil {
				return
			}

			keyBytes, err = ioutil.ReadFile("server.key")
			return
		}()
		if err != nil {
			return err
		}

		// See also:
		// `kubectl create secret tls tls-echo --cert=./server.crt --key=./server.key -o yaml --dry-run=client`
		objs = append(objs, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: deploy.GetName(),
			},
			Type: corev1.SecretTypeTLS,
			Data: map[string][]byte{
				"tls.crt": certBytes,
				"tls.key": keyBytes,
			},
		})

		return nil
	}()

	if err != nil {
		return fmt.Errorf("doing the thing: %w", err)
	}
	// END CUSTOM LOGIC

	dst := os.Stdout
	var typer runtime.ObjectTyper = scheme.Scheme

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
		if runtime.IsNotRegisteredError(err) {
			// This'd be snazzier if it suggested `blahv1.AddToScheme`
			return fmt.Errorf("scheme didn't recognize object of type %T (did you forget the relevant group's `AddToScheme`?): %w", obj, err)
		} else if err != nil {
			return fmt.Errorf("setting type info for %T: %w", obj, err)
		} else if len(types) != 1 {
			return fmt.Errorf("confused by the number of versions (%d) for struct %T", len(types), obj)
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
