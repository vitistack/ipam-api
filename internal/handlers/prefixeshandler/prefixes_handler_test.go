package prefixeshandler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterPrefix(t *testing.T) {
	type args struct {
		ginContext *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterPrefix(tt.args.ginContext)
		})
	}
}
