package cmd

import (
	"testing"
)

func Test_exportApplication(t *testing.T) {
	type args struct {
		exportlocation    string
		serviceProviderID string
		fileType          string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "",
			args: args{"/Users/indeewariwijesiri/Documents/Servers/GO", "e4412b35-bfea-4b22-8e05-4f6ff95498b4", "application/yaml"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exportApplication(tt.args.exportlocation, tt.args.serviceProviderID, tt.args.fileType); got != tt.want {
				t.Errorf("exportApplication() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_retreiveApplications(t *testing.T) {
	type args struct {
		reqUrl string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := retreiveApplications(tt.args.reqUrl); got != tt.want {
				t.Errorf("retreiveApplications() = %v, want %v", got, tt.want)
			}
		})
	}
}
