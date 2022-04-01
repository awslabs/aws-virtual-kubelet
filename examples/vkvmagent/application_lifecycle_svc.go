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
	"fmt"
	"log"
	"math/rand"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"

	pb "github.com/aws/aws-virtual-kubelet/proto/vkvmagent/v0"
)

type applicationLifecycleServer struct {
	pb.UnimplementedApplicationLifecycleServer
}

func (a *applicationLifecycleServer) LaunchApplication(
	ctx context.Context, request *pb.LaunchApplicationRequest) (*pb.LaunchApplicationResponse, error) {
	log.Printf("LaunchApplication invoked: %v", request)
	// TODO implement LaunchApplication behavior here
	log.Printf("Pod size is %d", request.GetPod().Size())

	return &pb.LaunchApplicationResponse{}, nil
}

func (a *applicationLifecycleServer) TerminateApplication(
	ctx context.Context, request *pb.TerminateApplicationRequest) (*pb.TerminateApplicationResponse, error) {
	log.Printf("TerminateApplication invoked: %v", request)
	// TODO implement TerminateApplication behavior here

	return &pb.TerminateApplicationResponse{}, nil
}

func (a *applicationLifecycleServer) CheckApplicationHealth(
	ctx context.Context, request *pb.ApplicationHealthRequest) (*pb.ApplicationHealthResponse, error) {
	log.Printf("CheckApplicationHealth invoked: %v", request)
	// TODO implement CheckApplicationHealth behavior here

	return &pb.ApplicationHealthResponse{
		PodStatus: happyPodStatus("🌈"),
	}, nil
}

func (a *applicationLifecycleServer) WatchApplicationHealth(
	request *pb.ApplicationHealthRequest, stream pb.ApplicationLifecycle_WatchApplicationHealthServer) error {
	log.Printf("WatchApplicationHealth invoked: %v", request)
	// TODO implement actual WatchApplicationHealth behavior here (example below send fake/random responses)

	var sleepTime int
	var appResponse *pb.ApplicationHealthResponse

	// start a goroutine with a random sleep that periodically sends a status
	//go func() {
	err := func() error {
		log.Printf("Starting watch sending loop...")
		for {
			sleepTime = rand.Intn(30)
			log.Printf("Sleeping for %d seconds...", sleepTime)
			time.Sleep(time.Duration(sleepTime) * time.Second)
			appResponse = &pb.ApplicationHealthResponse{
				PodStatus: happyPodStatus(fmt.Sprintf("Slept for %d seconds...", sleepTime)),
			}
			log.Printf("Sending app response: %+v", appResponse)
			if err := stream.Send(appResponse); err != nil {

				log.Printf("Error sending app response: %v", err)
				return err
			}
		}
	}()
	if err != nil {
		log.Printf("Something silly happened: %v", err)
	}
	//}()

	return nil
}

func happyPodStatus(message string) *corev1.PodStatus {
	happyConditions := []corev1.PodCondition{
		{
			Type:               corev1.PodScheduled,
			Status:             corev1.ConditionTrue,
			LastProbeTime:      metav1.Time{},
			LastTransitionTime: metav1.Time{},
			Message:            "🗓",
		},
		{
			Type:               corev1.ContainersReady,
			Status:             corev1.ConditionTrue,
			LastProbeTime:      metav1.Time{},
			LastTransitionTime: metav1.Time{},
			Message:            "🏺",
		},
		{
			Type:               corev1.PodInitialized,
			Status:             corev1.ConditionTrue,
			LastProbeTime:      metav1.Time{},
			LastTransitionTime: metav1.Time{},
			Message:            "༄",
		},
		{
			Type:               corev1.PodReady,
			Status:             corev1.ConditionTrue,
			LastProbeTime:      metav1.Time{},
			LastTransitionTime: metav1.Time{},
			Message:            "🤓",
		},
	}

	happyPodStatus := &corev1.PodStatus{
		Phase:      corev1.PodRunning,
		Conditions: happyConditions,
		Message:    message,
		ContainerStatuses: []corev1.ContainerStatus{{
			Name: "sample.container",
			State: corev1.ContainerState{
				Running: &corev1.ContainerStateRunning{
					StartedAt: metav1.Time{
						Time: time.Now(),
					},
				},
			},
			Ready:        true,
			RestartCount: 0,
			Image:        "image",
			ImageID:      "image-id",
			ContainerID:  "container-id",
		}},
	}

	return happyPodStatus
}
