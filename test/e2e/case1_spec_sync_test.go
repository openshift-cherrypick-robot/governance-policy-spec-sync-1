// Copyright (c) 2020 Red Hat, Inc.

package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-cluster-management/governance-policy-propagator/test/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const case1PolicyName string = "default.case1-test-policy"
const case1PolicyYaml string = "../resources/case1_spec_sync/case1-test-policy.yaml"

var _ = Describe("Test spec sync", func() {
	BeforeEach(func() {
		By("Creating a policy on hub cluster in ns:" + testNamespace)
		utils.Kubectl("apply", "-f", case1PolicyYaml, "-n", testNamespace,
			"--kubeconfig=../../kubeconfig_hub")
		plc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case1PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(plc).NotTo(BeNil())
	})
	AfterEach(func() {
		By("Deleting a policy on hub cluster in ns:" + testNamespace)
		utils.Kubectl("delete", "-f", case1PolicyYaml, "-n", testNamespace,
			"--kubeconfig=../../kubeconfig_hub")
		opt := metav1.ListOptions{}
		utils.ListWithTimeout(clientHubDynamic, gvrPolicy, opt, 0, true, defaultTimeoutSeconds)
		utils.ListWithTimeout(clientManagedDynamic, gvrPolicy, opt, 0, true, defaultTimeoutSeconds)
	})
	It("should create policy on managed cluster", func() {
		plc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case1PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(plc).NotTo(BeNil())
	})
	It("should update policy on managed cluster", func() {
		By("Patching " + case1PolicyYaml + " on hub with spec.remediationAction = enforce")
		hubPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, case1PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(hubPlc).NotTo(BeNil())
		Expect(hubPlc.Object["spec"].(map[string]interface{})["remediationAction"]).To(Equal("inform"))
		hubPlc.Object["spec"].(map[string]interface{})["remediationAction"] = "enforce"
		hubPlc, err := clientHubDynamic.Resource(gvrPolicy).Namespace(testNamespace).Update(hubPlc, metav1.UpdateOptions{})
		Expect(err).To(BeNil())
		Eventually(func() interface{} {
			managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case1PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return managedPlc.Object["spec"]
		}, defaultTimeoutSeconds, 1).Should(utils.SemanticEqual(hubPlc.Object["spec"]))
	})
	It("should update policy to a different policy template", func() {
		By("Updating policy on hub with ../resources/case1_propagation/case1-test-policy2.yaml")
		utils.Kubectl("apply",
			"-f", "../resources/case1_spec_sync/case1-test-policy2.yaml",
			"-n", testNamespace, "--kubeconfig=../../kubeconfig_hub")
		hubPlc := utils.GetWithTimeout(clientHubDynamic, gvrPolicy, case1PolicyName, testNamespace, true, defaultTimeoutSeconds)
		Expect(hubPlc).NotTo(BeNil())
		yamlPlc := utils.ParseYaml("../resources/case1_spec_sync/case1-test-policy2.yaml")
		Eventually(func() interface{} {
			managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrPolicy, case1PolicyName, testNamespace, true, defaultTimeoutSeconds)
			return managedPlc.Object["spec"]
		}, defaultTimeoutSeconds, 1).Should(utils.SemanticEqual(yamlPlc.Object["spec"]))
	})
	It("should delete policy on managed cluster", func() {
		By("Deleting policy on hub")
		utils.Kubectl("delete", "policy", "-n", testNamespace,
			"--all", "--kubeconfig=../../kubeconfig_hub")
		opt := metav1.ListOptions{}
		utils.ListWithTimeout(clientHubDynamic, gvrPolicy, opt, 0, true, defaultTimeoutSeconds)
		utils.ListWithTimeout(clientManagedDynamic, gvrPolicy, opt, 0, true, defaultTimeoutSeconds)
	})
})
