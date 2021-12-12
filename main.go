package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	newRootCMD().Execute()
}

func newRootCMD() *cobra.Command {
	var k8sOverrides clientcmd.ConfigOverrides
	cmd := cobra.Command{
		Use:  "halyard [filenames...]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p := newProcessor()
			for _, filename := range args {
				f, err := os.Open(filename)
				if err != nil {
					fatal(err)
				}
				defer f.Close()
				err = p.ReadResources(f, fileFormat(filepath.Ext(filename)))
				if err != nil {
					fatal(err)
				}

				configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
					clientcmd.NewDefaultClientConfigLoadingRules(),
					&k8sOverrides,
				)

				config, err := configLoader.ClientConfig()
				if err != nil {
					fatal(err)
				}

				resources, err := p.RenderResources()
				if err != nil {
					fatal(err)
				}

				for _, u := range resources {
					err = Apply(cmd.Context(), config, u)
					if err != nil {
						fatal(err)
					}
				}
				fmt.Println("Done")
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&k8sOverrides.CurrentContext, "context", "C", "", "Context override")

	// TODO: Enable namespace overrides
	// cmd.Flags().StringVarP(&k8sOverrides.Context.Namespace, "namespace", "n", "", "Namespace override")
	return &cmd
}

func fatal(err error) {
	fmt.Println("Error:", err)
	os.Exit(1)
}
