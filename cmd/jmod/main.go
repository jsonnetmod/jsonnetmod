package main

import (
	"context"

	"github.com/go-logr/logr"
)

func main() {
	ctx := logr.NewContext(context.Background(), log)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Error(err, "execute failed")
	}
}
