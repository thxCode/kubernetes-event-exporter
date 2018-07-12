package pipes

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDecode(t *testing.T) {
	ts := metav1.NewTime(time.Now())

	a := &HuaWeiEventLogBson2{
		Event: &v1.Event{
			TypeMeta: metav1.TypeMeta{
				Kind:       "event",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test",
				CreationTimestamp: metav1.NewTime(time.Now()),
				DeletionTimestamp: &ts,
				Annotations: map[string]string{
					"a": "b",
				},
				Finalizers: []string{"a","b"},
			},
		},
		EventName: "test",
	}

	b, err := a.MarshalBSONDocument()
	if err != nil {
		panic(err)
	}

	fmt.Println(b.ToExtJSON(true))
}
