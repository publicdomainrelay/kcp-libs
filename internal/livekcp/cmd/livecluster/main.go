package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/publicdomainrelay/kcp-libs/internal/livekcp"
)

func main() {
	envFile := flag.String("env", "", "file to write the environment to once the cluster is up")
	flag.Parse()
	if *envFile == "" {
		fmt.Fprintln(os.Stderr, "livecluster: --env is required")
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cluster, err := livekcp.Start(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "livecluster:", err)
		os.Exit(1)
	}
	defer cluster.Stop()

	body := fmt.Sprintf("export KCP_LIBS_KUBECONFIG=%q\nexport KCP_LIBS_SERVER=%q\n", cluster.Kubeconfig, cluster.Server)
	if err := os.WriteFile(*envFile+".tmp", []byte(body), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "livecluster:", err)
		os.Exit(1)
	}
	if err := os.Rename(*envFile+".tmp", *envFile); err != nil {
		fmt.Fprintln(os.Stderr, "livecluster:", err)
		os.Exit(1)
	}
	fmt.Println(cluster.Trace())
	<-ctx.Done()
}
