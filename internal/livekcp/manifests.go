package livekcp

import "syscall"

func procAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}

func workspaceYAML(name string) string {
	return `apiVersion: tenancy.kcp.io/v1alpha1
kind: Workspace
metadata:
  name: ` + name + `
spec:
  type:
    name: universal
    path: root
`
}

const SchemaYAML = `apiVersion: apis.kcp.io/v1alpha1
kind: APIResourceSchema
metadata:
  name: v1alpha1-1.probes.example.computer
spec:
  group: example.computer
  names:
    kind: Probe
    listKind: ProbeList
    plural: probes
    singular: probe
  scope: Namespaced
  versions:
    - name: v1alpha1
      served: true
      storage: true
      subresources:
        status: {}
      schema:
        type: object
        properties:
          apiVersion:
            type: string
          kind:
            type: string
          metadata:
            type: object
          spec:
            type: object
            properties:
              steps:
                type: integer
                format: int32
          status:
            type: object
            properties:
              phase:
                type: string
                enum:
                  - Pending
                  - Running
                  - Succeeded
                  - Failed
              observed:
                type: integer
                format: int32
              conditions:
                type: array
                items:
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
`

const ExportYAML = `apiVersion: apis.kcp.io/v1alpha2
kind: APIExport
metadata:
  name: ` + Export + `
spec:
  resources:
    - group: ` + Group + `
      name: ` + Resource + `
      schema: v1alpha1-1.probes.example.computer
      storage:
        crd: {}
`

const BindingYAML = `apiVersion: apis.kcp.io/v1alpha2
kind: APIBinding
metadata:
  name: ` + Export + `
spec:
  reference:
    export:
      path: ` + "root:" + ProviderWorkspace + `
      name: ` + Export + `
`
