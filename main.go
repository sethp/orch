package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/cli-runtime/pkg/resource"
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

	err := fs.WalkDir(manifest, ".", func(filename string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil
		}

		f, err := manifest.Open(filename)
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

	objs, err := builder.Do().Infos()
	if err != nil {
		return fmt.Errorf("collecting resources: %w", err)
	}

	// BEGIN CUSTOM LOGIC
	err = func() error {
		var cm *unstructured.Unstructured
		for _, obj := range objs {
			o := obj.Object.(*unstructured.Unstructured)
			if o.GetKind() == "ConfigMap" {
				cm = o
			}
		}

		if cm == nil {
			return errors.New("couldn't find configmap")
		}

		delete(cm.Object, "wrongField")
		cm.Object["data"] = map[string]string{
			"key": "value",
		}

		return nil
	}()

	if err != nil {
		return fmt.Errorf("doing the thing: %w", err)
	}
	// END CUSTOM LOGIC

	dst := os.Stdout
	switch Args.Output {
	case "json":
		serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil,
			json.SerializerOptions{
				Pretty: true,
			})
		for _, obj := range objs {
			err := serializer.Encode(obj.Object, dst)
			fmt.Fprintln(dst)
			if err != nil {
				return err
			}
		}
		return nil
	case "yaml":
		serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil,
			json.SerializerOptions{
				Yaml: true,
			})
		for _, obj := range objs {
			fmt.Fprintln(dst, "---")
			err := serializer.Encode(obj.Object, dst)
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

//go:embed manifest
var manifest embed.FS
