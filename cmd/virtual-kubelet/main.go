/*
This sample, non-production-ready code contains a Virtual Kubelet EC2-based provider and example VM Agent implementation.
© 2021 Amazon Web Services, Inc. or its affiliates. All Rights Reserved.

This AWS Content is provided subject to the terms of the AWS Customer Agreement
available at http://aws.amazon.com/agreement or other written agreement between
Customer and either Amazon Web Services, Inc. or Amazon Web Services EMEA SARL or both.
*/

package main

import (
	"context"
	"runtime"
	"strings"

	"github.com/aws/aws-virtual-kubelet/internal/config"
	"github.com/aws/aws-virtual-kubelet/internal/k8sutils"

	"github.com/virtual-kubelet/node-cli/provider"

	"github.com/virtual-kubelet/node-cli/opts"

	"github.com/aws/aws-virtual-kubelet/internal/ec2provider"

	"github.com/sirupsen/logrus"
	cli "github.com/virtual-kubelet/node-cli"
	logruscli "github.com/virtual-kubelet/node-cli/logrus"
	"github.com/virtual-kubelet/virtual-kubelet/log"
	logruslogger "github.com/virtual-kubelet/virtual-kubelet/log/logrus"
)

var (
	buildVersion = "N/A"
	buildTime    = "N/A"
	k8sVersion   = "v1.19.10" // This should follow the version of k8s.io/client-go we are importing
)

//nolint:funlen
func main() {
	ctx := cli.ContextWithCancelOnSignal(context.Background())

	// configure CLI logging
	logger := logrus.StandardLogger()
	log.L = logruslogger.FromLogrus(logrus.NewEntry(logger))
	logConfig := &logruscli.Config{LogLevel: "info"}

	// set options from ENV
	o, err := opts.FromEnv()
	if err != nil {
		log.G(ctx).Fatal(err)
	}

	// set default and generated option values (some can be overridden by CLI flags)
	o.Provider = "ec2"
	o.ProviderConfigPath = config.DefaultConfigLocation
	o.Version = strings.Join([]string{k8sVersion, "vk-aws-ec2", buildVersion, runtime.GOARCH}, "-")

	log.G(ctx).Infof("Virtual Kubelet (%v) starting...", o.Version)

	var p *ec2provider.Ec2Provider

	// instantiate CLI command
	cliCommand, err := cli.New(ctx,
		cli.WithBaseOpts(o),
		cli.WithCLIVersion(o.Version, buildTime),
		cli.WithPersistentFlags(logConfig.FlagSet()),
		// Set the NodeName option to ensure that the ENI-based node name matches the VK internal node name
		// ⚠️ The VK internal node name _must_ be set correctly or CreatePod is never called.
		// `WithPersistentPreRunCallback` adds a callback which is called _after_ flags are processed
		// but _before_ running the command or any sub-command (at which point VK/k8s starts calling Ec2Provider fxns).
		cli.WithPersistentPreRunCallback(func() error {
			// NOTE ⚠️ Provider instantiation needs to happen here (after CLI flag processing but _before_ virtual
			//  kubelet runs the CLI command). This is because VK creates a KubeInformer filtered by NodeName to watch
			//  for pod activity, _then_ runs the  Init function from WithProvider. At that point VK will have already
			//  used the (likely incorrect) node name and will never see pod-related provider activity (and will never
			//  call PodLifecycleHandler functions).  *This case is difficult to notice* because the node name _appears_
			//  correct in all logging and queries. Since our provider generates the node name dynamically from an ENI
			//  IP address, we _must_ instantiate the provider before that KubeInformer filter is created and override
			//  the NodeName CLI option to ensure it is set _before_ the KubeInformer is created. See links for details.
			//
			//  KubeInformer creation (line 111 of virtual-kubelet/node-cli root.go):
			//  https://github.com/virtual-kubelet/node-cli/blob/bfe728730f54651b279d6e19e92599ede4e6fa1c/internal/commands/root/root.go#L111
			//
			//  Provider instantiation "callback" function call (line 163 of virtual-kubelet/node-cli root.go):
			//  https://github.com/virtual-kubelet/node-cli/blob/bfe728730f54651b279d6e19e92599ede4e6fa1c/internal/commands/root/root.go#L163

			// create temporary initial config
			initConfig := provider.InitConfig{
				ConfigPath: o.ProviderConfigPath,
				//NodeName:          o.NodeName, // ignoring NodeName since it will be set explicitly by provider
				OperatingSystem:   o.OperatingSystem,
				DaemonPort:        o.ListenPort,
				KubeClusterDomain: o.KubeClusterDomain,
			}

			extendedConfig := config.ExtendedConfig{
				KubeConfigPath: o.KubeConfigPath,
			}

			p, err = ec2provider.NewEc2Provider(ctx, initConfig, extendedConfig)
			if err != nil {
				log.G(ctx).Fatal(err)
			}

			// rehydrate cache prior to starting provider (or k8s will just ask us _right back_ for the list)
			k8sClient, err := k8sutils.NewK8sClient(o.KubeConfigPath)
			podList, err := k8sClient.GetPods(context.TODO(), p.NodeName)
			if err != nil {
				return err
			}
			podCache := ec2provider.NewPodCache()
			// populate pod cache from k8s pod list
			podCache.Populate(podList)
			// update provider so it starts with the cache pre-loaded (before k8s asks us for it)
			p.PopulateCache(podCache)

			// warn users if we have overridden a custom node name provided via CLI
			if o.NodeName != "" {
				log.G(ctx).Warnf(
					"Ignoring '--nodename %v' argument (node name is set by the provider and cannot be overridden)",
					o.NodeName,
				)
			}

			// update node name from provider property (*must* match generated name and *must* be set here)
			o.NodeName = p.NodeName

			return logruscli.Configure(logConfig, logger)
		}),
		cli.WithProvider("ec2", func(cfg provider.InitConfig) (provider.Provider, error) {
			// NOTE this being safe depends on the fact that the function passed to `WithPersistentPreRunCallback` above
			//  will be called _before_ this code passed in the function to `WithProvider` here.  This behavior is
			//  unlikely to change in VK, but worth noting.
			return p, nil
		}),
	)
	// handle possible cliCommand creation errors
	if err != nil {
		log.G(ctx).Fatal(err)
	}

	// run the CLI command which starts VK processing / handling
	if err := cliCommand.Run(ctx); err != nil {
		log.G(ctx).Fatal(err)
	}
}
