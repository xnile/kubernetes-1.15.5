package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	namespace := "default"

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	//待创建的文件
	f, err := os.Open("xnile.yaml")

	if err != nil {
		log.Fatal(err)
	}
	d := yaml.NewYAMLOrJSONDecoder(f, 4096)
	dc := client.Discovery()

	restMapperRes, err := restmapper.GetAPIGroupResources(dc)
	if err != nil {
		log.Fatal(err)
	}

	restMapper := restmapper.NewDiscoveryRESTMapper(restMapperRes) //

	for {

		ext := runtime.RawExtension{}
		if err := d.Decode(&ext); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		// fmt.Println("raw: ", string(ext.Raw))

		// runtime.Object
		obj, gvk, err := unstructured.UnstructuredJSONScheme.Decode(ext.Raw, nil, nil)
		fmt.Printf("gvk: %+v\n", gvk)

		mapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		// fmt.Printf("mapping:%+v\n", mapping)
		if err != nil {
			log.Fatal(err)
		}

		// runtime.Object转换为unstructed
		unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("unstructuredObj: %+v", unstructuredObj)

		var unstruct unstructured.Unstructured

		unstruct.Object = unstructuredObj

		if md, ok := unstruct.Object["metadata"]; ok {
			metadata := md.(map[string]interface{})
			if internalns, ok := metadata["namespace"]; ok {
				namespace = internalns.(string)
			}
		}

		// 动态客户端
		dclient, err := dynamic.NewForConfig(config)
		if err != nil {
			log.Fatal(err)
		}

		res, err := dclient.Resource(mapping.Resource).Namespace(namespace).Create(&unstruct, metav1.CreateOptions{})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res)

	}
}
