package main

import (
	"context"
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"sort"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

//go:embed index.html
var indexHtml string

type App struct {
	client kubernetes.Interface
}

func newApp() *App {
	config, err := getConfig()
	if err != nil {
		panic(err.Error())
	}

	client, err := newClient(config)
	if err != nil {
		panic(err.Error())
	}
	return &App{
		client: client,
	}
}

type IngressInfo struct {
	Host string
}

func main() {
	app := newApp()
	http.HandleFunc("/", app.handler)
	listenAddr := ":8080"
	log.Printf("listning at %s", listenAddr)
	http.ListenAndServe(listenAddr, nil)
}

func getConfig() (*rest.Config, error) {
	log.Println("trying InClusterConfig")
	config, err := rest.InClusterConfig()
	if err == nil {
		log.Println("loaded config successfully")
		return config, nil
	}
	log.Println(err)

	log.Printf("trying %s\n", clientcmd.RecommendedHomeFile)
	config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err == nil {
		log.Println("loaded config successfully")
		return config, nil
	}
	log.Println(err)

	return nil, err

}

func newClient(config *rest.Config) (kubernetes.Interface, error) {
	return kubernetes.NewForConfig(config)
}

func getIngressList(ctx context.Context, client kubernetes.Interface) ([]IngressInfo, error) {
	ingresses, err := client.NetworkingV1().Ingresses("").List(ctx, meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var infos = []IngressInfo{}
	for _, ing := range ingresses.Items {
		for _, rule := range ing.Spec.Rules {
			infos = append(infos, IngressInfo{Host: rule.Host})
		}
	}
	return infos, nil
}

func (app *App) handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ingresses, err := getIngressList(ctx, app.client)
	if err != nil {
		panic(err.Error())
	}
	sort.Slice(ingresses, func(i, j int) bool { return ingresses[i].Host < ingresses[j].Host })

	tmpl := template.Must(template.New("index").Parse(indexHtml))
	data := ingresses
	err = tmpl.Execute(w, data)
	if err != nil {
		panic(err.Error())
	}
}
