package health

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_setPodHandler(t *testing.T) {
	type args struct {
		pod     *corev1.Pod
		handler *CheckHandler
	}

	providedHandler := &CheckHandler{
		in: make(chan *checkResult),
	}

	tests := []struct {
		name string
		args args
		want *CheckHandler
	}{
		{
			name: "Missing handler annotation uses provided handler function",
			args: args{
				pod: &corev1.Pod{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name:        "pod-no-annotation",
						Annotations: nil,
					},
					Spec:   corev1.PodSpec{},
					Status: corev1.PodStatus{},
				},
				handler: providedHandler,
			},
			want: providedHandler,
		},
		// TODO(guicejg): add tests for pod handler annotation cases
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setPodHandler(tt.args.pod, tt.args.handler); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("setPodHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
