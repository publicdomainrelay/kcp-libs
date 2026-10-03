package kcp

import "strings"

const (
	RootWorkspace = "root"

	ClusterAnnotation = "kcp.io/cluster"

	PathAnnotation = "kcp.io/path"

	DefaultServiceDomain = "kcp.local"

	ServiceSegment = "svc"
)

func ServiceLabels(logicalCluster string) string {
	parts := strings.Split(logicalCluster, ":")
	out := make([]string, 0, len(parts))
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == "" || parts[i] == RootWorkspace {
			continue
		}
		out = append(out, parts[i])
	}
	return strings.Join(out, ".")
}

func ServiceFQDN(name, namespace, logicalCluster, domain string) string {
	labels := ServiceLabels(logicalCluster)
	if namespace == "" {
		namespace = "default"
	}
	host := name + "." + namespace
	if labels != "" {
		host += "." + labels
	}
	return host + "." + ServiceSegment + "." + domain
}

func ServiceSuffix(domain string) string {
	return "." + ServiceSegment + "." + domain
}

func ClusterFromLabels(labels string) string {
	parts := strings.Split(labels, ".")
	out := []string{RootWorkspace}
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			out = append(out, parts[i])
		}
	}
	return strings.Join(out, ":")
}

func SplitServiceFQDN(host, domain string) (name, namespace, logicalCluster string, ok bool) {
	suffix := ServiceSuffix(domain)
	if !strings.HasSuffix(host, suffix) {
		return "", "", "", false
	}
	bits := strings.Split(strings.TrimSuffix(host, suffix), ".")
	if len(bits) < 2 {
		return "", "", "", false
	}
	return bits[0], bits[1], ClusterFromLabels(strings.Join(bits[2:], ".")), true
}
