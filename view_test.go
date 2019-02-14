package fox

import (
	"net/http"
	"testing"
)

// 后面再测
func TestModelRoute(t *testing.T) {
	type args struct {
		model      interface{}
		modelSlice interface{}
		disableApi []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"BaseTest", args{&Person{}, []Person{}, []string{}}, false},
		{"BaseTest", args{&Person{}, []Person{}, []string{http.MethodPost}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ModelRoute(tt.args.model, tt.args.modelSlice, tt.args.disableApi...); (err != nil) != tt.wantErr {
				t.Errorf("ModelRoute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
