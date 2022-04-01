package utils

import corev1 "k8s.io/api/core/v1"

var podNotifier func(*corev1.Pod)

func GetNotifier() func(*corev1.Pod) {
	return podNotifier
}

func SetNotifier(f func(*corev1.Pod)) {
	podNotifier = f
}
