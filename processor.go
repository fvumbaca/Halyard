package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/imdario/mergo"
	yaml "gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type fileFormat string

const (
	FormatYAML = fileFormat(".yaml")
)

const (
	layerAnnotation  = "halyard.sh/layer"
	defaultLayerName = "base"
)

type processor struct {
	Resources map[string]map[string]*unstructured.Unstructured
}

func newProcessor() *processor {
	return &processor{
		Resources: make(map[string]map[string]*unstructured.Unstructured),
	}
}

func (p *processor) addResource(u *unstructured.Unstructured) {
	layer, ok := u.GetAnnotations()[layerAnnotation]
	if !ok {
		layer = defaultLayerName
	}

	resources, ok := p.Resources[layer]
	if !ok {
		p.Resources[layer] = make(map[string]*unstructured.Unstructured)
		resources = p.Resources[layer]
	}
	resources[generateResourceID(u)] = u
}

func (p *processor) ReadResource(input io.Reader, format fileFormat) error {
	if format != FormatYAML {
		return errors.New("format not supported")
	}

	r := yaml.NewDecoder(input)

	var err error
	for err == nil {
		var obj unstructured.Unstructured
		err = r.Decode(&obj.Object)
		if err != nil {
			break
		}
		p.addResource(&obj)
	}

	if err == io.EOF {
		err = nil
	}

	return err
}

func (p *processor) ReadResourceFiles(filenames []string) error {
	for _, filename := range filenames {
		if err := p.ReadResourceFile(filename); err != nil {
			return err
		}
	}
	return nil
}

func (p *processor) ReadResourceFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return p.ReadResource(f, fileFormat(filepath.Ext(filename)))
}

func generateResourceID(u *unstructured.Unstructured) string {
	return u.GetKind() + "/" + u.GetNamespace() + "/" + u.GetName()
}

func (p *processor) RenderResources() ([]*unstructured.Unstructured, error) {
	m := make(map[string]*unstructured.Unstructured)

	for _, layer := range p.Layers() {
		fmt.Println("Processing layer", layer)
		resources := p.Resources[layer]
		for rid, o := range resources {
			if _, ok := m[rid]; ok {
				mergo.Merge(m[rid], o, mergo.WithOverride)
			} else {
				m[rid] = o
			}
		}
	}

	var r []*unstructured.Unstructured
	for _, v := range m {
		r = append(r, v)
	}
	return r, nil
}

func (p *processor) Layers() []string {
	var ls []string
	for l := range p.Resources {
		ls = append(ls, l)
	}
	sort.Slice(ls, func(i, j int) bool {
		if ls[i] == "base" {
			return true
		}
		if ls[j] == "base" {
			return false
		}
		return sort.StringSlice(ls).Less(i, j)
	})
	return ls
}
