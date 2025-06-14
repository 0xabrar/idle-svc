package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// typed clients
	discovery "k8s.io/client-go/kubernetes/typed/discovery/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	// core API object for Service
	coreapi "k8s.io/api/core/v1"

	"k8s.io/client-go/tools/clientcmd"
)

/* ---------- data model ---------- */

type orphan struct {
	Namespace string `json:"namespace"`
	Service   string `json:"service"`
	Type      string `json:"type"`
	ClusterIP string `json:"clusterIP"`
	Age       string `json:"age"`
}

/* ---------- main ---------- */

func main() {
	// ----- CLI flags -----
	allNS := flag.Bool("A", false, "scan all namespaces")
	nsFlag := flag.String("namespace", "", "scan a single namespace")
	jsonOut := flag.Bool("json", false, "output JSON instead of table")
	exitErr := flag.Bool("exit-code", false, "exit 1 when idle services found")
	flag.Parse()

	if *allNS && *nsFlag != "" {
		fatal("", fmt.Errorf("use either -A or --namespace, not both"))
	}

	// ----- kubeconfig -----
	cfg, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		fatal("kubeconfig", err)
	}

	// ----- typed clients -----
	coreClient, err := corev1.NewForConfig(cfg)
	if err != nil {
		fatal("core client", err)
	}
	discoClient, err := discovery.NewForConfig(cfg)
	if err != nil {
		fatal("discovery client", err)
	}

	ns := *nsFlag // empty string = current default namespace

	// ----- list Services -----
	svcs, err := coreClient.Services(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fatal("list services", err)
	}

	var orphans []orphan

	for _, svc := range svcs.Items {
		// EndpointSlice label: kubernetes.io/service-name=<service-name>
		slices, err := discoClient.EndpointSlices(svc.Namespace).List(
			context.TODO(),
			metav1.ListOptions{LabelSelector: "kubernetes.io/service-name=" + svc.Name},
		)

		readyAddrs := 0
		if err == nil {
			for _, sl := range slices.Items {
				for _, ep := range sl.Endpoints {
					if ep.Conditions.Ready != nil && *ep.Conditions.Ready {
						readyAddrs++
					}
				}
			}
		} else {
			// Fallback to legacy Endpoints API (older clusters)
			ep, err2 := coreClient.Endpoints(svc.Namespace).Get(context.TODO(), svc.Name, metav1.GetOptions{})
			if err2 == nil && len(ep.Subsets) > 0 {
				for _, ss := range ep.Subsets {
					readyAddrs += len(ss.Addresses)
				}
			}
		}

		if readyAddrs == 0 {
			orphans = append(orphans, mapSvc(svc))
		}
	}

	// ----- output -----
	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(orphans)
	} else if len(orphans) == 0 {
		fmt.Println("👍  no idle services found")
	} else {
		tbl := tablewriter.NewWriter(os.Stdout)
		tbl.SetHeader([]string{"NAMESPACE", "SERVICE", "TYPE", "CLUSTER-IP", "AGE"})
		for _, o := range orphans {
			tbl.Append([]string{o.Namespace, o.Service, o.Type, o.ClusterIP, o.Age})
		}
		tbl.Render()
	}

	if *exitErr && len(orphans) > 0 {
		os.Exit(1)
	}
}

/* ---------- helpers ---------- */

func mapSvc(svc coreapi.Service) orphan {
	age := time.Since(svc.CreationTimestamp.Time).Round(time.Minute)
	return orphan{
		Namespace: svc.Namespace,
		Service:   svc.Name,
		Type:      string(svc.Spec.Type),
		ClusterIP: svc.Spec.ClusterIP,
		Age:       age.String(),
	}
}

func fatal(context string, err error) {
	if context != "" {
		fmt.Fprintf(os.Stderr, "%s: %v\n", context, err)
	} else {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	os.Exit(2)
}

