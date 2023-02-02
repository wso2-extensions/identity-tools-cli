package cmd

import "testing"

func Test_importApplication(t *testing.T) {
	type args struct {
		importFilePath string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "import",
			args: args{"/Users/indeewariwijesiri/Documents/Servers/GO/dc9f20ec-75c6-41b4-bb30-e82f93413c08_export.xml"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := importApplication(tt.args.importFilePath); got != tt.want {
				t.Errorf("importApplication() = %v, want %v", got, tt.want)
			}
		})
	}
}
