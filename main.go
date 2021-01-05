package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/labstack/echo/v4"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	addr          = flag.String("listen-address", ":8000", "The metrics endpoint port")
	refreshPeriod = getInt(getEnv("REFRESH_PERIOD", "120"))
	namespace     = getEnv("NAMESPACE", "default")
	namespaces    = getNamespaces(getEnv("NAMESPACES", "default"))
	endpoint      = getEnv("PROMETHEUS_ENDPOINT", "/metrics")
	kubeconfig    = getEnv("KUBE_CONFIG", "kubeconfig")
)

var (
	bxServiceInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bx_service_info",
			Help: "Version information about BX Services",
		},
		[]string{"namespace", "app", "pod", "version", "images"},
	)
)

func getNamespaces(list string) []string {
	return strings.Split(list, ",")
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func getInt(s string) int {
	refreshPeriod, err := strconv.Atoi(s)
	if err != nil {
		return 60
	}
	return refreshPeriod
}

func init() {
	prometheus.MustRegister(bxServiceInfo)
	log.SetOutput(os.Stdout)
}

// MetricData is the version metrics struct
type MetricData struct {
	Namespace string `json:"namespace"`
	App       string `json:"app"`
	Pod       string `json:"pod"`
	Version   string `json:"version"`
	Images    string `json:"images"`
}

func getMetricData(clientset *kubernetes.Clientset, namespace string) []MetricData {
	metricData := []MetricData{}
	metricData = nil

	pods, _ := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	for _, pod := range pods.Items {

		containers := pod.Spec.Containers
		metadata := pod.ObjectMeta
		namespace := metadata.Namespace
		app := metadata.Labels["app"]
		pod := metadata.Name
		version := metadata.Labels["version"]

		//Long winded way first
		imageArray := []string{}
		for _, container := range containers {
			imageArray = append(imageArray, container.Image)
		}
		images := strings.Join(imageArray, ", ")

		data := MetricData{
			Namespace: namespace,
			App:       app,
			Pod:       pod,
			Version:   version,
			Images:    images,
		}

		metricData = append(metricData, data)
	}
	return metricData
}

func main() {

	log.Info(fmt.Sprintf("Starting version metrics service with %d second refresh", refreshPeriod))
	log.Info(fmt.Sprintf("Watching the %s namespaces", namespaces))
	log.Info(fmt.Sprintf("Prometheus endpoint enabled at %s", endpoint))

	//config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	//if err != nil {
	//	panic(err)
	//}
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	metrics := []MetricData{}

	go func() {
		for {
			metrics = nil
			for _, namespace := range namespaces {
				log.Info("Refreshing metrics...", namespace)

				data := getMetricData(clientset, namespace)
				metrics = append(metrics, data...)

			}

			for _, metric := range metrics {
				bxServiceInfo.With(
					prometheus.Labels{
						"namespace": metric.Namespace,
						"app":       metric.App,
						"pod":       metric.Pod,
						"version":   metric.Version,
						"images":    metric.Images}).Set(1)
			}

			time.Sleep(time.Second * time.Duration(refreshPeriod))

		}
	}()

	go func() {
		e := echo.New()

		e.GET("/", func(c echo.Context) error {
			return c.JSONPretty(http.StatusOK, &metrics, "  ")
		})
		e.Logger.Fatal(e.Start(":9000"))
	}()

	http.Handle(endpoint, promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))
	log.Fatal(http.ListenAndServe(*addr, nil))

}
