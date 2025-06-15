package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/0xabrar/idle-svc/pkg/orphanfinder"

	discovery "k8s.io/client-go/kubernetes/typed/discovery/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/tools/clientcmd"
)

/* ---------- main ---------- */

func main() {
	// ----- CLI flags -----
	allNS := flag.Bool("A", false, "scan all namespaces")
	nsFlag := flag.String("namespace", "", "scan a single namespace")
	jsonOut := flag.Bool("json", false, "output JSON instead of table")
	exitErr := flag.Bool("exit-code", false, "exit 1 when idle services found")
	listen := flag.String("listen", "", "HTTP listen address for Prometheus metrics (e.g. :9090); empty disables metrics server")
	watch := flag.Bool("watch", false, "continuously rescan every 30s (implies metrics if --listen is set)")
	interval := flag.Duration("interval", 30*time.Second, "scan interval for --watch mode")
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

	ctx := context.TODO()

	if *listen != "" {
		initMetrics(*listen)
	}

	scanOnce := func() []orphanfinder.Orphan {
		orphans, err := orphanfinder.Find(ctx, coreClient, discoClient, *nsFlag, *allNS)
		if err != nil {
			fatal("scan", err)
		}
		if *listen != "" {
			updateMetrics(orphans)
		}
		return orphans
	}

	output := func(orphans []orphanfinder.Orphan) {
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

	if *watch {
		for {
			o := scanOnce()
			output(o)
			time.Sleep(*interval)
		}
	}

	// one-shot
	output(scanOnce())
}

/* ---------- helpers ---------- */

func fatal(context string, err error) {
	if context != "" {
		fmt.Fprintf(os.Stderr, "%s: %v\n", context, err)
	} else {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	os.Exit(2)
}

