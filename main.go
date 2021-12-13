package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	_, err := newRootCMD().ExecuteContextC(context.TODO())
	if err != nil {
		fatal(err)
	}
}

func newRootCMD() *cobra.Command {
	cmd := cobra.Command{
		Use: "halyard",
	}
	cmd.AddCommand(
		newApplyCMD(),
		newYAMLCMD(),
	)
	return &cmd
}

func newApplyCMD() *cobra.Command {
	var k8sOverrides clientcmd.ConfigOverrides
	cmd := cobra.Command{
		Use:  "Apply [filenames...]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				clientcmd.NewDefaultClientConfigLoadingRules(),
				&k8sOverrides,
			)

			config, err := configLoader.ClientConfig()
			if err != nil {
				fatal(err)
			}

			p := newProcessor()
			err = p.ReadResourceFiles(args)
			if err != nil {
				fatal(err)
			}

			resources, err := p.RenderResources()
			if err != nil {
				fatal(err)
			}
			err = Apply(cmd.Context(), config, resources)
			if err != nil {
				fatal(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&k8sOverrides.CurrentContext, "context", "C", "", "Context override")

	// TODO: Enable namespace overrides
	// cmd.Flags().StringVarP(&k8sOverrides.Context.Namespace, "namespace", "n", "", "Namespace override")
	return &cmd
}

func newYAMLCMD() *cobra.Command {
	cmd := cobra.Command{
		Use:  "yaml [filenames...]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := newProcessor()
			err := p.ReadResourceFiles(args)
			if err != nil {
				fatal(err)
			}

			resources, err := p.RenderResources()
			if err != nil {
				fatal(err)
			}
			err = Template(os.Stdout, resources)
			if err != nil {
				fatal(err)
			}
			return nil
		},
	}

	return &cmd
}

func fatal(err error) {
	fmt.Println("Error:", err)
	os.Exit(1)
}
