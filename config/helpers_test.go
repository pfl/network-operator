// Copyright 2026 Intel Corporation. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helpers

import (
	"testing"
)

// malformedYaml cannot be parsed as yaml at all: the flow sequence is never
// closed.
const malformedYaml = `
metadata:
  name: [unterminated
`

// mistypedYaml parses as yaml, but the value types do not match the target
// Kubernetes object.
const mistypedYaml = `
metadata: "this should be an object"
`

// expectPanic runs fn and fails the test unless fn panics.
func expectPanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected a panic, got none")
		}
	}()

	fn()
}

func TestGetDaemonSet(t *testing.T) {
	content := []byte(`
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: test-daemonset
  namespace: test-namespace
  labels:
    app: test
spec:
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      hostNetwork: true
      serviceAccountName: test-serviceaccount
      containers:
        - name: test-container
          image: test-image:latest
`)

	ds := GetDaemonSet(content)
	if ds == nil {
		t.Fatal("expected to receive a valid daemonset")
	}

	if ds.Kind != "DaemonSet" {
		t.Errorf("expected kind to be 'DaemonSet', got: %s", ds.Kind)
	}

	if ds.Name != "test-daemonset" {
		t.Errorf("expected name to be 'test-daemonset', got: %s", ds.Name)
	}

	if ds.Namespace != "test-namespace" {
		t.Errorf("expected namespace to be 'test-namespace', got: %s", ds.Namespace)
	}

	if ds.Labels["app"] != "test" {
		t.Errorf("expected label app to be 'test', got: %s", ds.Labels["app"])
	}

	if !ds.Spec.Template.Spec.HostNetwork {
		t.Error("expected host network to be enabled")
	}

	if ds.Spec.Template.Spec.ServiceAccountName != "test-serviceaccount" {
		t.Errorf("expected service account name to be 'test-serviceaccount', got: %s",
			ds.Spec.Template.Spec.ServiceAccountName)
	}

	if len(ds.Spec.Template.Spec.Containers) != 1 {
		t.Fatalf("expected 1 container, got: %d", len(ds.Spec.Template.Spec.Containers))
	}

	if ds.Spec.Template.Spec.Containers[0].Name != "test-container" {
		t.Errorf("expected container name to be 'test-container', got: %s",
			ds.Spec.Template.Spec.Containers[0].Name)
	}

	if ds.Spec.Template.Spec.Containers[0].Image != "test-image:latest" {
		t.Errorf("expected container image to be 'test-image:latest', got: %s",
			ds.Spec.Template.Spec.Containers[0].Image)
	}
}

func TestGetDaemonSetEmptyContent(t *testing.T) {
	ds := GetDaemonSet([]byte{})
	if ds == nil {
		t.Fatal("expected to receive an empty but valid daemonset")
	}

	if ds.Name != "" {
		t.Errorf("expected an empty name, got: %s", ds.Name)
	}
}

func TestGetDaemonSetInvalidContent(t *testing.T) {
	for name, content := range map[string]string{
		"malformed": malformedYaml,
		"mistyped":  mistypedYaml,
	} {
		t.Run(name, func(t *testing.T) {
			expectPanic(t, func() {
				GetDaemonSet([]byte(content))
			})
		})
	}
}

func TestGetServiceAccount(t *testing.T) {
	content := []byte(`
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-serviceaccount
  namespace: test-namespace
`)

	sa := GetServiceAccount(content)
	if sa == nil {
		t.Fatal("expected to receive a valid service account")
	}

	if sa.Kind != "ServiceAccount" {
		t.Errorf("expected kind to be 'ServiceAccount', got: %s", sa.Kind)
	}

	if sa.Name != "test-serviceaccount" {
		t.Errorf("expected name to be 'test-serviceaccount', got: %s", sa.Name)
	}

	if sa.Namespace != "test-namespace" {
		t.Errorf("expected namespace to be 'test-namespace', got: %s", sa.Namespace)
	}
}

func TestGetServiceAccountEmptyContent(t *testing.T) {
	sa := GetServiceAccount([]byte{})
	if sa == nil {
		t.Fatal("expected to receive an empty but valid service account")
	}

	if sa.Name != "" {
		t.Errorf("expected an empty name, got: %s", sa.Name)
	}
}

func TestGetServiceAccountInvalidContent(t *testing.T) {
	for name, content := range map[string]string{
		"malformed": malformedYaml,
		"mistyped":  mistypedYaml,
	} {
		t.Run(name, func(t *testing.T) {
			expectPanic(t, func() {
				GetServiceAccount([]byte(content))
			})
		})
	}
}

func TestGetClusterRole(t *testing.T) {
	content := []byte(`
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: test-clusterrole
rules:
  - apiGroups:
      - ""
    resources:
      - nodes
    verbs:
      - get
      - list
  - apiGroups:
      - resource.k8s.io
    resources:
      - resourceclaims/driver
    resourceNames:
      - test.driver
    verbs:
      - update
`)

	cr := GetClusterRole(content)
	if cr == nil {
		t.Fatal("expected to receive a valid cluster role")
	}

	if cr.Kind != "ClusterRole" {
		t.Errorf("expected kind to be 'ClusterRole', got: %s", cr.Kind)
	}

	if cr.Name != "test-clusterrole" {
		t.Errorf("expected name to be 'test-clusterrole', got: %s", cr.Name)
	}

	if len(cr.Rules) != 2 {
		t.Fatalf("expected 2 rules, got: %d", len(cr.Rules))
	}

	if len(cr.Rules[0].Resources) != 1 || cr.Rules[0].Resources[0] != "nodes" {
		t.Errorf("expected the first rule to cover 'nodes', got: %v", cr.Rules[0].Resources)
	}

	if len(cr.Rules[0].Verbs) != 2 {
		t.Errorf("expected 2 verbs in the first rule, got: %v", cr.Rules[0].Verbs)
	}

	if len(cr.Rules[1].ResourceNames) != 1 || cr.Rules[1].ResourceNames[0] != "test.driver" {
		t.Errorf("expected the second rule to name 'test.driver', got: %v", cr.Rules[1].ResourceNames)
	}
}

func TestGetClusterRoleEmptyContent(t *testing.T) {
	cr := GetClusterRole([]byte{})
	if cr == nil {
		t.Fatal("expected to receive an empty but valid cluster role")
	}

	if len(cr.Rules) != 0 {
		t.Errorf("expected no rules, got: %d", len(cr.Rules))
	}
}

func TestGetClusterRoleInvalidContent(t *testing.T) {
	for name, content := range map[string]string{
		"malformed": malformedYaml,
		"mistyped":  mistypedYaml,
	} {
		t.Run(name, func(t *testing.T) {
			expectPanic(t, func() {
				GetClusterRole([]byte(content))
			})
		})
	}
}

func TestGetRoleBinding(t *testing.T) {
	content := []byte(`
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: test-rolebinding
  namespace: test-namespace
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: test-role
subjects:
  - kind: ServiceAccount
    name: test-serviceaccount
    namespace: test-namespace
`)

	rb := GetRoleBinding(content)
	if rb == nil {
		t.Fatal("expected to receive a valid role binding")
	}

	if rb.Kind != "RoleBinding" {
		t.Errorf("expected kind to be 'RoleBinding', got: %s", rb.Kind)
	}

	if rb.Name != "test-rolebinding" {
		t.Errorf("expected name to be 'test-rolebinding', got: %s", rb.Name)
	}

	if rb.Namespace != "test-namespace" {
		t.Errorf("expected namespace to be 'test-namespace', got: %s", rb.Namespace)
	}

	if rb.RoleRef.Kind != "Role" || rb.RoleRef.Name != "test-role" {
		t.Errorf("expected the role ref to point to Role/test-role, got: %s/%s",
			rb.RoleRef.Kind, rb.RoleRef.Name)
	}

	if len(rb.Subjects) != 1 {
		t.Fatalf("expected 1 subject, got: %d", len(rb.Subjects))
	}

	if rb.Subjects[0].Kind != "ServiceAccount" || rb.Subjects[0].Name != "test-serviceaccount" {
		t.Errorf("expected the subject to be ServiceAccount/test-serviceaccount, got: %s/%s",
			rb.Subjects[0].Kind, rb.Subjects[0].Name)
	}
}

func TestGetRoleBindingEmptyContent(t *testing.T) {
	rb := GetRoleBinding([]byte{})
	if rb == nil {
		t.Fatal("expected to receive an empty but valid role binding")
	}

	if len(rb.Subjects) != 0 {
		t.Errorf("expected no subjects, got: %d", len(rb.Subjects))
	}
}

func TestGetRoleBindingInvalidContent(t *testing.T) {
	for name, content := range map[string]string{
		"malformed": malformedYaml,
		"mistyped":  mistypedYaml,
	} {
		t.Run(name, func(t *testing.T) {
			expectPanic(t, func() {
				GetRoleBinding([]byte(content))
			})
		})
	}
}

func TestGetClusterRoleBinding(t *testing.T) {
	content := []byte(`
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: test-clusterrolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: test-clusterrole
subjects:
  - kind: ServiceAccount
    name: test-serviceaccount
    namespace: kube-system
`)

	crb := GetClusterRoleBinding(content)
	if crb == nil {
		t.Fatal("expected to receive a valid cluster role binding")
	}

	if crb.Kind != "ClusterRoleBinding" {
		t.Errorf("expected kind to be 'ClusterRoleBinding', got: %s", crb.Kind)
	}

	if crb.Name != "test-clusterrolebinding" {
		t.Errorf("expected name to be 'test-clusterrolebinding', got: %s", crb.Name)
	}

	if crb.RoleRef.Kind != "ClusterRole" || crb.RoleRef.Name != "test-clusterrole" {
		t.Errorf("expected the role ref to point to ClusterRole/test-clusterrole, got: %s/%s",
			crb.RoleRef.Kind, crb.RoleRef.Name)
	}

	if len(crb.Subjects) != 1 {
		t.Fatalf("expected 1 subject, got: %d", len(crb.Subjects))
	}

	if crb.Subjects[0].Namespace != "kube-system" {
		t.Errorf("expected the subject namespace to be 'kube-system', got: %s",
			crb.Subjects[0].Namespace)
	}
}

func TestGetClusterRoleBindingEmptyContent(t *testing.T) {
	crb := GetClusterRoleBinding([]byte{})
	if crb == nil {
		t.Fatal("expected to receive an empty but valid cluster role binding")
	}

	if len(crb.Subjects) != 0 {
		t.Errorf("expected no subjects, got: %d", len(crb.Subjects))
	}
}

func TestGetClusterRoleBindingInvalidContent(t *testing.T) {
	for name, content := range map[string]string{
		"malformed": malformedYaml,
		"mistyped":  mistypedYaml,
	} {
		t.Run(name, func(t *testing.T) {
			expectPanic(t, func() {
				GetClusterRoleBinding([]byte(content))
			})
		})
	}
}

func TestGetDeviceClass(t *testing.T) {
	content := []byte(`
apiVersion: resource.k8s.io/v1
kind: DeviceClass
metadata:
  name: test-deviceclass
spec:
  selectors:
    - cel:
        expression: device.driver == "test.driver"
    - cel:
        expression: device.attributes["test.driver"].rdma == true
`)

	dc := GetDeviceClass(content)
	if dc == nil {
		t.Fatal("expected to receive a valid device class")
	}

	if dc.Kind != "DeviceClass" {
		t.Errorf("expected kind to be 'DeviceClass', got: %s", dc.Kind)
	}

	if dc.Name != "test-deviceclass" {
		t.Errorf("expected name to be 'test-deviceclass', got: %s", dc.Name)
	}

	if len(dc.Spec.Selectors) != 2 {
		t.Fatalf("expected 2 selectors, got: %d", len(dc.Spec.Selectors))
	}

	for i, expected := range []string{
		`device.driver == "test.driver"`,
		`device.attributes["test.driver"].rdma == true`,
	} {
		cel := dc.Spec.Selectors[i].CEL
		if cel == nil {
			t.Errorf("expected selector %d to have a CEL selector", i)

			continue
		}

		if cel.Expression != expected {
			t.Errorf("expected selector %d expression to be %q, got: %q", i, expected, cel.Expression)
		}
	}
}

func TestGetDeviceClassEmptyContent(t *testing.T) {
	dc := GetDeviceClass([]byte{})
	if dc == nil {
		t.Fatal("expected to receive an empty but valid device class")
	}

	if len(dc.Spec.Selectors) != 0 {
		t.Errorf("expected no selectors, got: %d", len(dc.Spec.Selectors))
	}
}

func TestGetDeviceClassInvalidContent(t *testing.T) {
	for name, content := range map[string]string{
		"malformed": malformedYaml,
		"mistyped":  mistypedYaml,
	} {
		t.Run(name, func(t *testing.T) {
			expectPanic(t, func() {
				GetDeviceClass([]byte(content))
			})
		})
	}
}
