/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package ec2provider

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/aws/aws-virtual-kubelet/internal/vkvmaclient"

	"github.com/aws/aws-virtual-kubelet/internal/health"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-virtual-kubelet/internal/config"

	vkvmagent_v0 "github.com/aws/aws-virtual-kubelet/proto/vkvmagent/v0"

	"github.com/aws/aws-virtual-kubelet/internal/metrics"

	"github.com/aws/aws-virtual-kubelet/internal/utils"

	"github.com/virtual-kubelet/node-cli/manager"

	"k8s.io/klog/v2"

	"github.com/virtual-kubelet/node-cli/provider"

	"github.com/virtual-kubelet/virtual-kubelet/errdefs"

	"github.com/virtual-kubelet/virtual-kubelet/node/api"
	"github.com/virtual-kubelet/virtual-kubelet/node/api/statsv1alpha1"

	corev1 "k8s.io/api/core/v1"
)

// Ec2Provider implements PodLifecycleHandler which defines the interface used by the PodController to react to new and
//  changed pods scheduled to the node that is being managed.
// Errors produced by these methods should implement an interface from
//  github.com/virtual-kubelet/virtual-kubelet/errdefs package in order for the core logic to be able to understand
//  the type of failure.
// See https://pkg.go.dev/github.com/virtual-kubelet/virtual-kubelet/node
type Ec2Provider struct {
	NodeName string
	EniNode  *EniNode

	rm                 *manager.ResourceManager
	internalIP         string
	daemonEndpointPort int32
	pods               *PodCache
	startTime          time.Time
	podNotifier        func(*corev1.Pod)
	computeManager     *computeManager
	podMonitor         *health.PodMonitor
	checkHandler       *health.CheckHandler
	warmPool           *WarmPoolManager
}

func NewEc2Provider(ctx context.Context, cfg provider.InitConfig, extCfg config.ExtendedConfig) (*Ec2Provider, error) {
	updateVerbosityLevel()

	klog.V(1).Infof("📣 Verbosity level one (1) logging enabled!")
	klog.V(2).Infof("🚨 Verbosity level two (2) logging enabled!")

	klog.Infof("Creating EC2 Provider with initial config '%+v'", cfg)

	var configLoader config.Loader = &config.FileLoader{ConfigFilePath: cfg.ConfigPath}

	err := config.InitConfig(configLoader)
	if err != nil {
		klog.ErrorS(err, "Can't process config")
	}

	p := Ec2Provider{
		rm: cfg.ResourceManager,
	}

	p.EniNode, err = GetOrCreateEniNode(ctx)
	if err != nil {
		klog.Errorf(
			"Unable to get or create node (ENI).  Are the AWS credentials expired/invalid? %v",
			err,
		)
		return nil, err
	}
	// set the provider local nodeName property
	p.NodeName = p.EniNode.name

	p.warmPool, err = NewWarmPool(ctx, &p)
	if err != nil {
		panic("handle warm pool instantiation error")
	}
	p.warmPool.fillAndMaintain()

	//p.warmPool, err = NewV2WarmPool(ctx, &p)

	p.computeManager, err = NewComputeManager(ctx)
	if err != nil {
		panic("handle compute manager instantiation error")
	}

	p.checkHandler = health.NewCheckHandler(p.podNotifier)

	// start metrics endpoint
	go metrics.ExposeMetrics()

	return &p, nil
}

// NOTE VK Interface methods/functions should generally return a standard error to VK/k8s.  This allows VK/k8s own retry
//  and backoff mechanisms to take over.  Cases where a VK errdefs error makes sense are exceptions (e.g. GetPod()).
//  see github.com/virtual-kubelet/virtual-kubelet/errdefs for details

// PodLifecycleHandler methods (required)
// See https://pkg.go.dev/github.com/virtual-kubelet/virtual-kubelet/node#PodLifecycleHandler

func (p *Ec2Provider) CreatePod(ctx context.Context, pod *corev1.Pod) error {
	klog.Infof("Received CreatePod request for pod %v(%v)", pod.Name, pod.Namespace)

	// create (but don't start) pod monitor
	podMonitor, err := health.NewPodMonitor(pod, p.checkHandler)
	if err != nil {
		klog.ErrorS(err, "Can't create pod monitor", "pod", klog.KObj(pod))
		return err
	}

	// add pod to cache
	metaPod := NewMetaPod(pod, podMonitor, p.podNotifier)
	p.pods.Set(utils.GetPodCacheKey(pod.Namespace, pod.Name), metaPod)

	// launch EC2
	instanceID, privateIP, err := p.computeManager.GetCompute(ctx, p, pod)
	if err != nil {
		klog.ErrorS(err, "Error getting compute for pod", "pod", klog.KObj(pod))
		return err
	}

	pod.Status.PodIP = privateIP
	pod.Status.HostIP = privateIP

	// launch application
	// NOTE LaunchApplicationResponse is currently empty (so we discard it)
	_, err = p.launchApplication(ctx, pod)
	if err != nil {
		klog.ErrorS(err, "Error launching application", "pod", klog.KObj(pod))
		return err
	}

	if utils.PodIsWarmPool(ctx, pod, len(p.warmPool.config)) {
		// mark the instance as IN_USE
		err = p.warmPool.updateEC2Tags(ctx, instanceID, "set_in_use", *pod)
		if err != nil {
			klog.ErrorS(err, "Can't update EC2 tags for Warm Pool", "instance",
				instanceID, pod, "pod", klog.KObj(pod))
			return err
		}
	}

	// start monitoring
	podMonitor.Start(ctx)

	// notify k8s with pod status update
	p.podNotifier(pod)

	// increment metric
	metrics.PodsLaunched.Inc()

	return nil
}

func (p *Ec2Provider) UpdatePod(ctx context.Context, pod *corev1.Pod) error {
	klog.Infof("Received UpdatePod request for pod %v(%v)", pod.Name, pod.Namespace)

	// update pod cache
	podKey := utils.GetPodCacheKey(pod.Namespace, pod.Name)
	err := p.pods.UpdatePod(podKey, pod)
	if err != nil {
		klog.ErrorS(err, "Error updating pod cache", "pod", klog.KObj(pod))
		return err
	}

	klog.Infof("Updated pod %v(%v)", pod.Name, pod.Namespace)

	return nil
}

func (p *Ec2Provider) DeletePod(ctx context.Context, pod *corev1.Pod) error {
	klog.Infof("Received DeletePod request for pod %v(%v)", pod.Name, pod.Namespace)

	podKey := utils.GetPodCacheKey(pod.Namespace, pod.Name)
	metaPod := p.pods.Get(podKey)

	var err error

	// stop monitoring
	err = p.stopPodMonitor(ctx, metaPod)
	if err != nil {
		klog.ErrorS(err, "Could not stop pod monitoring", "pod", klog.KObj(metaPod.pod))
		return err
	}

	// terminate application
	err = p.terminateApp(ctx, metaPod)
	if err != nil {
		klog.ErrorS(err, "Could not terminate application", "pod", klog.KObj(metaPod.pod))
		return err
	}

	// terminate EC2
	err = p.computeManager.DeleteCompute(ctx, p, pod)
	if err != nil {
		klog.Errorf("Error deleting compute: %v", err)
		return err
	}

	// notify k8s
	p.notifyPodDelete(pod)

	// delete from cache
	p.pods.Delete(podKey)

	klog.InfoS("Pod deleted", "pod", klog.KObj(metaPod.pod))

	return nil
}

func (p *Ec2Provider) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	klog.Infof("Received GetPod request for %v(%v)", name, namespace)

	metaPod := p.pods.Get(utils.GetPodCacheKey(namespace, name))
	if metaPod == nil {
		return nil, errdefs.NotFoundf("Pod %v(%v) does not exist", name, namespace)
	}
	return metaPod.pod, nil
}

func (p *Ec2Provider) GetPodStatus(ctx context.Context, namespace, name string) (*corev1.PodStatus, error) {
	klog.Infof("Received GetPodStatus request for %v(%v)", name, namespace)

	metaPod := p.pods.Get(utils.GetPodCacheKey(namespace, name))
	if metaPod == nil {
		return nil, errdefs.NotFoundf("Pod %v(%v) does not exist", name, namespace)
	}
	return &metaPod.pod.Status, nil
}

func (p *Ec2Provider) GetPods(ctx context.Context) ([]*corev1.Pod, error) {
	klog.Info("Received GetPods request.")

	podList := p.pods.GetPodList()
	klog.Infof("podList length is %v", len(podList))

	return podList, nil
}

// PodNotifier methods (recommended)
// See https://pkg.go.dev/github.com/virtual-kubelet/virtual-kubelet/node#PodNotifier

func (p *Ec2Provider) NotifyPods(ctx context.Context, f func(*corev1.Pod)) {
	klog.Info("NotifyPods notifier callback function set")
	p.podNotifier = f
	utils.SetNotifier(f)
}

// Provider methods (required)
// See https://github.com/virtual-kubelet/virtual-kubelet/blob/master/cmd/virtual-kubelet/internal/provider/provider.go
// NOTE The requirement to implement this method is _not_ currently captured in the documentation for virtual-kubelet
//  https://pkg.go.dev/github.com/virtual-kubelet/virtual-kubelet#section-documentation

func (p *Ec2Provider) ConfigureNode(ctx context.Context, node *corev1.Node) {
	// NOTE ConfigureNode can't return an error because the VK interface doesn't include one in the method signature
	klog.InfoS("Configuring node", "node", klog.KObj(node))
	_, err := p.EniNode.Configure(ctx, node)
	if err != nil {
		klog.ErrorS(err, "Error configuring node", "node", klog.KObj(node))
	}
}

// Provider methods (optional)
// See https://pkg.go.dev/github.com/virtual-kubelet/virtual-kubelet/node/nodeutil#Provider

func (p *Ec2Provider) GetContainerLogs(ctx context.Context, namespace, podName, containerName string, opts api.ContainerLogOpts) (io.ReadCloser, error) {
	return nil, errdefs.NotFound("Not yet implemented")
}

func (p *Ec2Provider) RunInContainer(ctx context.Context, namespace, podName, containerName string, cmd []string, attach api.AttachIO) error {
	return errdefs.NotFound("Not yet implemented")
}

func (p *Ec2Provider) GetStatsSummary(ctx context.Context) (*statsv1alpha1.Summary, error) {
	return nil, errdefs.NotFound("Not yet implemented")
}

func (p *Ec2Provider) launchApplication(
	ctx context.Context, pod *corev1.Pod) (*vkvmagent_v0.LaunchApplicationResponse, error) {

	cfg := config.Config()

	var launchAppResp *vkvmagent_v0.LaunchApplicationResponse

	// TODO(guicejg): 🚨 Get port from config
	vkvmaClient := vkvmaclient.NewVkvmaClient(pod.Status.PodIP, cfg.VKVMAgentConnectionConfig.Port)

	appClient, err := vkvmaClient.GetApplicationLifecycleClient(ctx)
	if err != nil {
		klog.ErrorS(err, "Error getting ApplicationLifecycleClient", "pod", klog.KObj(pod))

		err2 := p.computeManager.DeleteCompute(ctx, p, pod)
		if err2 != nil {
			klog.ErrorS(err2, "Error deleting compute while cleaning up failed CreatePod", "original error", err)
		}
		p.pods.Delete(utils.GetPodCacheKey(pod.Namespace, pod.Name))

		return nil, err
	}

	launchAppResp, err = appClient.LaunchApplication(ctx, &vkvmagent_v0.LaunchApplicationRequest{
		Pod: pod,
	})
	if err != nil {
		klog.ErrorS(err, "Error Launching Application for pod", "pod", klog.KObj(pod))

		err2 := p.computeManager.DeleteCompute(ctx, p, pod)
		if err2 != nil {
			klog.ErrorS(err2, "Error deleting compute while cleaning up failed CreatePod", "original error", err)
		}
		p.pods.Delete(utils.GetPodCacheKey(pod.Namespace, pod.Name))

		return nil, err
	}
	return launchAppResp, nil
}

// NOTE while klog flags are present in the command line options presented by `virtual-kubelet -h`, it does not process
//  them correctly for some reason, which means flags like ` --klog.v Level` are silently discarded. 😑
// updateVerbosityLevel explicitly sets the klog verbosity from the intended commandline arg to work around a bug in
//  node-cli
func updateVerbosityLevel() {
	var level klog.Level

	// check command line args for klog verbosity arg and update verbosity if found
	for i, arg := range os.Args {
		if arg == "--klog.v" {
			// ignoring errors here preserves the original behavior of silently failing to set the log level
			_ = level.Set(os.Args[i+1])
		}
	}
}

func (p *Ec2Provider) deletePodSkipApp(ctx context.Context, pod *corev1.Pod) {
	klog.InfoS("EC2 failure...recreating pod", "pod", klog.KObj(pod))

	podKey := utils.GetPodCacheKey(pod.Namespace, pod.Name)
	metaPod := p.pods.Get(podKey)

	var err error

	// stop monitoring
	err = p.stopPodMonitor(ctx, metaPod)
	if err != nil {
		klog.ErrorS(err, "Could not stop pod monitoring", "pod", klog.KObj(metaPod.pod))
	}

	// terminate EC2
	err = p.computeManager.DeleteCompute(ctx, p, pod)
	if err != nil {
		klog.Errorf("Error deleting compute: %v", err)
	}

	// notify k8s
	p.notifyPodDelete(pod)

	// delete from cache
	p.pods.Delete(podKey)
}

// handlePodStatusUpdate is called by pod monitors to update pod status
func (p *Ec2Provider) handlePodStatusUpdate(ctx context.Context, pod *corev1.Pod, podStatus corev1.PodStatus) error {
	klog.Infof("Pod %v(%v) received a status update: %+v", pod.Name, pod.Namespace, podStatus)

	pod.Status = podStatus
	p.podNotifier(pod)

	return nil
}

func (p *Ec2Provider) stopPodMonitor(ctx context.Context, metaPod *MetaPod) error {
	// stop pod monitor goroutine (if it exists)
	var err error

	if metaPod != nil && metaPod.monitor != nil {
		metaPod.monitor.Stop(ctx)
	} else {
		err = errors.New("metaPod or metaPod.monitor is nil")
	}
	return err
}

func (p *Ec2Provider) terminateApp(ctx context.Context, metaPod *MetaPod) error {
	cfg := config.Config()

	var termAppResp *vkvmagent_v0.TerminateApplicationResponse

	pod := metaPod.pod

	vkvmaClient := vkvmaclient.NewVkvmaClient(pod.Status.PodIP, cfg.VKVMAgentConnectionConfig.Port)

	appClient, err := vkvmaClient.GetApplicationLifecycleClient(ctx)
	if err != nil {
		klog.ErrorS(err, "Could not get ApplicationLifecycleClient", "pod", klog.KObj(metaPod.pod))
		return err
	}

	termAppResp, err = appClient.TerminateApplication(ctx, &vkvmagent_v0.TerminateApplicationRequest{})
	if err != nil {
		klog.ErrorS(err, "Could not Terminate Application", "pod", klog.KObj(metaPod.pod))
		return err
	}

	klog.Infof("Application terminated with response: %+v", termAppResp)
	return nil
}

func (p *Ec2Provider) notifyPodDelete(pod *corev1.Pod) {
	pod.Status.Phase = corev1.PodSucceeded
	pod.Status.Reason = "ProviderPodDeleted"

	// set container statuses for termination
	for idx := range pod.Status.ContainerStatuses {
		pod.Status.ContainerStatuses[idx].Ready = false
		pod.Status.ContainerStatuses[idx].State = corev1.ContainerState{
			Terminated: &corev1.ContainerStateTerminated{
				Message:    "Pod deletion requested",
				FinishedAt: metav1.Now(),
			},
		}
	}
	p.podNotifier(pod)
}

// PopulateCache enables loading of pod cache from k8s itself prior to k8s asking us for the list of pods 😵‍💫
func (p *Ec2Provider) PopulateCache(cache *PodCache) {
	var err error

	metaPods := cache.GetList()

	klog.Infof("Populating cache: loading %v pods and creating monitors", len(metaPods))
	for _, metaPod := range metaPods {
		klog.Infof("Recreating pod monitor for pod %v(%v) (populated from cache)",
			metaPod.pod.Name, metaPod.pod.Namespace)

		handler := health.NewCheckHandler(p.podNotifier)
		metaPod.monitor, err = health.NewPodMonitor(metaPod.pod, handler)
		if err != nil {
			klog.Errorf("Can't create pod health monitor for pod %v(%v): %v",
				metaPod.pod.Name, metaPod.pod.Namespace, err)
		}
		metaPod.monitor.Start(context.TODO())
	}

	p.pods = cache
}
