// Copyright 2023 Terramate GmbH
// SPDX-License-Identifier: MPL-2.0

//go:build localhostEndpoints

package cli

import (
	"os"

	"github.com/hashicorp/go-uuid"
)

const cloudDefaultBaseURL = "http://localhost:3001"

func generateRunID() (string, error) {
	if runid := os.Getenv("TM_TEST_RUN_ID"); runid != "" {
		return runid, nil
	}
	return uuid.GenerateUUID()
}
