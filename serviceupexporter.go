package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"gopkg.in/yaml.v2"
	"os/exec"
	"strings"
	"time"
	"io/ioutil"
	"log"
	"os"
)

type conf struct {
	Services []Service `yaml:"services"`
}

type Service struct{
	ServiceName string `yaml:"servicename"`
	MetricName string `yaml:"metricname"`
	Help string `yaml:"help"`
}

func (c *conf) getConf() *conf {
	yamlFile, err := ioutil.ReadFile("servicenames.yml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}

func heartBeat(servicename string){
	//fmt.Println("Foo")
	out, err := exec.Command("service",servicename, "status").Output()
	if err != nil {
		//fmt.Println("ERR:", err.Error())
		out = []byte(err.Error())
	}else{
	//fmt.Printf("%s", out)
	}
	hostname, err :=os.Hostname()
	if strings.Contains(string(out), "running"){
		GaugeMap[servicename].WithLabelValues(hostname).Set(0)
	}else{
		GaugeMap[servicename].WithLabelValues(hostname).Set(1)
	}
	time.Sleep(time.Second * 5)
	heartBeat(servicename)
}

//Configuration object
var con *conf
//A map of Prometheus Gauges
var GaugeMap = make(map[string] *prometheus.GaugeVec)

func main() {
	con = new(conf)
	con.getConf()
	for _,element :=	 range con.Services {
		GaugeMap[element.ServiceName] = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name : element.MetricName, Help: element.Help}, []string{ "hostname" })
		prometheus.MustRegister(GaugeMap[element.ServiceName])
		go heartBeat(element.ServiceName)
	}

	fmt.Println("Starting Http server and listening on port 9105...")
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("starting heartbeat")
	if err := http.ListenAndServe("0.0.0.0:9105", nil); err != nil {
		fmt.Println("Failed to make connection" + err.Error())
	}
}