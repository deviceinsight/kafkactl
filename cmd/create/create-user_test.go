package create_test

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/deviceinsight/kafkactl/v5/internal/testutil"
)

func TestCreateUserIntegration(t *testing.T) {
	testutil.StartIntegrationTestWithContext(t, "sasl-admin")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	username := fmt.Sprintf("testuser-%d", time.Now().Unix())

	// Test creating user with SCRAM-SHA-256
	_, err := kafkaCtl.Execute("create", "user", username, "--password", "testpass123")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	output := kafkaCtl.GetStdOut()
	if !strings.Contains(output, fmt.Sprintf("user '%s' has been created", username)) {
		t.Fatalf("expected user creation message, got: %s", output)
	}

	// Verify user was created by describing it
	_, err = kafkaCtl.Execute("describe", "user", username)
	if err != nil {
		t.Fatalf("failed to describe user: %v", err)
	}

	output = kafkaCtl.GetStdOut()
	if !strings.Contains(output, username) {
		t.Fatalf("expected user in describe output, got: %s", output)
	}
	if !strings.Contains(output, "SCRAM-SHA-256") {
		t.Fatalf("expected SCRAM-SHA-256 mechanism in output, got: %s", output)
	}

	// Clean up
	_, err = kafkaCtl.Execute("delete", "user", username, "--mechanism", "SCRAM-SHA-256")
	if err != nil {
		t.Logf("cleanup failed (may be expected): %v", err)
	}
}

func TestCreateUserWithCustomSaltIntegration(t *testing.T) {
	testutil.StartIntegrationTestWithContext(t, "sasl-admin")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	username := fmt.Sprintf("testsaltuser-%d", time.Now().Unix())
	customSalt := base64.StdEncoding.EncodeToString([]byte("customsalt123456"))

	// Test creating user with custom salt
	_, err := kafkaCtl.Execute("create", "user", username,
		"--password", "testpass123",
		"--salt", customSalt,
		"--iterations", "8192")
	if err != nil {
		t.Fatalf("failed to create user with custom salt: %v", err)
	}

	output := kafkaCtl.GetStdOut()
	if !strings.Contains(output, fmt.Sprintf("user '%s' has been created", username)) {
		t.Fatalf("expected user creation message, got: %s", output)
	}

	// Verify user was created with correct iterations
	_, err = kafkaCtl.Execute("describe", "user", username)
	if err != nil {
		t.Fatalf("failed to describe user: %v", err)
	}

	output = kafkaCtl.GetStdOut()
	if !strings.Contains(output, "8192") {
		t.Fatalf("expected 8192 iterations in output, got: %s", output)
	}

	// Clean up
	_, err = kafkaCtl.Execute("delete", "user", username, "--mechanism", "SCRAM-SHA-256")
	if err != nil {
		t.Logf("cleanup failed (may be expected): %v", err)
	}
}

func TestCreateUserSCRAMSHA512Integration(t *testing.T) {
	testutil.StartIntegrationTestWithContext(t, "sasl-admin")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	username := fmt.Sprintf("testsha512user-%d", time.Now().Unix())

	// Test creating user with SCRAM-SHA-512
	_, err := kafkaCtl.Execute("create", "user", username,
		"--password", "testpass123",
		"--mechanism", "SCRAM-SHA-512")
	if err != nil {
		t.Fatalf("failed to create user with SCRAM-SHA-512: %v", err)
	}

	output := kafkaCtl.GetStdOut()
	if !strings.Contains(output, fmt.Sprintf("user '%s' has been created", username)) {
		t.Fatalf("expected user creation message, got: %s", output)
	}

	// Verify user was created with SCRAM-SHA-512
	_, err = kafkaCtl.Execute("describe", "user", username)
	if err != nil {
		t.Fatalf("failed to describe user: %v", err)
	}

	output = kafkaCtl.GetStdOut()
	if !strings.Contains(output, "SCRAM-SHA-512") {
		t.Fatalf("expected SCRAM-SHA-512 mechanism in output, got: %s", output)
	}

	// Clean up
	_, err = kafkaCtl.Execute("delete", "user", username, "--mechanism", "SCRAM-SHA-512")
	if err != nil {
		t.Logf("cleanup failed (may be expected): %v", err)
	}
}

func TestCreateUserMissingPasswordIntegration(t *testing.T) {
	testutil.StartIntegrationTestWithContext(t, "sasl-admin")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	username := fmt.Sprintf("testnopass-%d", time.Now().Unix())

	// Test creating user without password should fail
	_, err := kafkaCtl.Execute("create", "user", username)
	if err == nil {
		t.Fatalf("expected error when creating user without password")
	}

	if !strings.Contains(err.Error(), "password") {
		t.Fatalf("expected password required error, got: %v", err)
	}
}

func TestCreateUserInvalidMechanismIntegration(t *testing.T) {
	testutil.StartIntegrationTestWithContext(t, "sasl-admin")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	username := fmt.Sprintf("testinvalid-%d", time.Now().Unix())

	// Test creating user with invalid mechanism
	_, err := kafkaCtl.Execute("create", "user", username,
		"--password", "testpass123",
		"--mechanism", "INVALID-MECHANISM")
	if err == nil {
		t.Fatalf("expected error when creating user with invalid mechanism")
	}

	if !strings.Contains(err.Error(), "unsupported SCRAM mechanism") {
		t.Fatalf("expected unsupported mechanism error, got: %v", err)
	}
}

func TestAlterUserIntegration(t *testing.T) {
	testutil.StartIntegrationTestWithContext(t, "sasl-admin")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	username := fmt.Sprintf("testalter-%d", time.Now().Unix())

	// First create a user
	_, err := kafkaCtl.Execute("create", "user", username, "--password", "originalpass")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Now alter the user's password
	_, err = kafkaCtl.Execute("alter", "user", username, "--password", "newpass123")
	if err != nil {
		t.Fatalf("failed to alter user: %v", err)
	}

	output := kafkaCtl.GetStdOut()
	if !strings.Contains(output, fmt.Sprintf("user '%s' credentials have been updated", username)) {
		t.Fatalf("expected user update message, got: %s", output)
	}

	// Verify user still exists
	_, err = kafkaCtl.Execute("describe", "user", username)
	if err != nil {
		t.Fatalf("failed to describe user after alter: %v", err)
	}

	// Clean up
	_, err = kafkaCtl.Execute("delete", "user", username, "--mechanism", "SCRAM-SHA-256")
	if err != nil {
		t.Logf("cleanup failed (may be expected): %v", err)
	}
}

func TestDeleteUserIntegration(t *testing.T) {
	testutil.StartIntegrationTestWithContext(t, "sasl-admin")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	username := fmt.Sprintf("testdelete-%d", time.Now().Unix())

	// First create a user
	_, err := kafkaCtl.Execute("create", "user", username, "--password", "testpass123")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Verify user exists
	_, err = kafkaCtl.Execute("describe", "user", username)
	if err != nil {
		t.Fatalf("failed to describe user before delete: %v", err)
	}

	// Delete the user
	_, err = kafkaCtl.Execute("delete", "user", username, "--mechanism", "SCRAM-SHA-256")
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	output := kafkaCtl.GetStdOut()
	if !strings.Contains(output, fmt.Sprintf("user '%s' SCRAM-SHA-256 credentials have been deleted", username)) {
		t.Fatalf("expected user deletion message, got: %s", output)
	}

	// Verify user no longer exists (or has no SCRAM-SHA-256 credentials)
	_, err = kafkaCtl.Execute("describe", "user", username)
	if err == nil {
		// User might still exist with other mechanisms, check the output
		output = kafkaCtl.GetStdOut()
		if strings.Contains(output, "SCRAM-SHA-256") {
			t.Fatalf("user should not have SCRAM-SHA-256 credentials after deletion, got: %s", output)
		}
	}
	// If error occurred, that's also acceptable (user completely gone)
}

func TestGetUsersIntegration(t *testing.T) {
	testutil.StartIntegrationTestWithContext(t, "sasl-admin")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	// Test getting all users (might be empty)
	_, err := kafkaCtl.Execute("get", "users")
	if err != nil {
		t.Fatalf("failed to get users: %v", err)
	}

	// Create a test user to ensure we have at least one
	username := fmt.Sprintf("testgetuser-%d", time.Now().Unix())
	_, err = kafkaCtl.Execute("create", "user", username, "--password", "testpass123")
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Now get users again
	_, err = kafkaCtl.Execute("get", "users")
	if err != nil {
		t.Fatalf("failed to get users after creating test user: %v", err)
	}

	output := kafkaCtl.GetStdOut()
	if !strings.Contains(output, username) {
		t.Fatalf("expected to find test user in users list, got: %s", output)
	}

	// Test JSON output
	_, err = kafkaCtl.Execute("get", "users", "-o", "json")
	if err != nil {
		t.Fatalf("failed to get users in JSON format: %v", err)
	}

	output = kafkaCtl.GetStdOut()
	if !strings.Contains(output, `"name"`) || !strings.Contains(output, username) {
		t.Fatalf("expected JSON format with user data, got: %s", output)
	}

	// Clean up
	_, err = kafkaCtl.Execute("delete", "user", username, "--mechanism", "SCRAM-SHA-256")
	if err != nil {
		t.Logf("cleanup failed (may be expected): %v", err)
	}
}

func TestDescribeUserIntegration(t *testing.T) {
	testutil.StartIntegrationTestWithContext(t, "sasl-admin")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	username := fmt.Sprintf("testdescribe-%d", time.Now().Unix())

	// Create a user with both SCRAM mechanisms
	_, err := kafkaCtl.Execute("create", "user", username,
		"--password", "testpass123",
		"--mechanism", "SCRAM-SHA-256")
	if err != nil {
		t.Fatalf("failed to create user with SCRAM-SHA-256: %v", err)
	}

	_, err = kafkaCtl.Execute("create", "user", username,
		"--password", "testpass123",
		"--mechanism", "SCRAM-SHA-512")
	if err != nil {
		t.Fatalf("failed to create user with SCRAM-SHA-512: %v", err)
	}

	// Describe the user
	_, err = kafkaCtl.Execute("describe", "user", username)
	if err != nil {
		t.Fatalf("failed to describe user: %v", err)
	}

	output := kafkaCtl.GetStdOut()
	if !strings.Contains(output, username) {
		t.Fatalf("expected username in describe output, got: %s", output)
	}
	if !strings.Contains(output, "SCRAM-SHA-256") {
		t.Fatalf("expected SCRAM-SHA-256 mechanism in output, got: %s", output)
	}
	if !strings.Contains(output, "SCRAM-SHA-512") {
		t.Fatalf("expected SCRAM-SHA-512 mechanism in output, got: %s", output)
	}

	// Test YAML output
	_, err = kafkaCtl.Execute("describe", "user", username, "-o", "yaml")
	if err != nil {
		t.Fatalf("failed to describe user in YAML format: %v", err)
	}

	output = kafkaCtl.GetStdOut()
	if !strings.Contains(output, "name:") || !strings.Contains(output, "mechanisms:") {
		t.Fatalf("expected YAML format with user data, got: %s", output)
	}

	// Clean up
	_, err = kafkaCtl.Execute("delete", "user", username, "--mechanism", "SCRAM-SHA-256")
	if err != nil {
		t.Logf("cleanup SHA-256 failed (may be expected): %v", err)
	}
	_, err = kafkaCtl.Execute("delete", "user", username, "--mechanism", "SCRAM-SHA-512")
	if err != nil {
		t.Logf("cleanup SHA-512 failed (may be expected): %v", err)
	}
}

func TestDescribeNonExistentUserIntegration(t *testing.T) {
	testutil.StartIntegrationTestWithContext(t, "sasl-admin")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	nonExistentUser := fmt.Sprintf("nonexistent-%d", time.Now().Unix())

	// Try to describe a user that doesn't exist
	_, err := kafkaCtl.Execute("describe", "user", nonExistentUser)
	if err == nil {
		t.Fatalf("expected error when describing non-existent user")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected 'not found' error, got: %v", err)
	}
}

func TestUserLifecycleIntegration(t *testing.T) {
	testutil.StartIntegrationTestWithContext(t, "sasl-admin")
	kafkaCtl := testutil.CreateKafkaCtlCommand()

	username := fmt.Sprintf("lifecycle-%d", time.Now().Unix())

	// 1. Create user
	_, err := kafkaCtl.Execute("create", "user", username, "--password", "pass1")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// 2. Verify in user list
	_, err = kafkaCtl.Execute("get", "users")
	if err != nil {
		t.Fatalf("failed to get users: %v", err)
	}
	if !strings.Contains(kafkaCtl.GetStdOut(), username) {
		t.Fatalf("user not found in users list")
	}

	// 3. Update password
	_, err = kafkaCtl.Execute("alter", "user", username, "--password", "newpass2")
	if err != nil {
		t.Fatalf("failed to alter user: %v", err)
	}

	// 4. Add second mechanism
	_, err = kafkaCtl.Execute("create", "user", username,
		"--password", "pass3",
		"--mechanism", "SCRAM-SHA-512")
	if err != nil {
		t.Fatalf("failed to add SCRAM-SHA-512: %v", err)
	}

	// 5. Verify both mechanisms exist
	_, err = kafkaCtl.Execute("describe", "user", username)
	if err != nil {
		t.Fatalf("failed to describe user: %v", err)
	}
	output := kafkaCtl.GetStdOut()
	if !strings.Contains(output, "SCRAM-SHA-256") || !strings.Contains(output, "SCRAM-SHA-512") {
		t.Fatalf("expected both mechanisms, got: %s", output)
	}

	// 6. Delete one mechanism
	_, err = kafkaCtl.Execute("delete", "user", username, "--mechanism", "SCRAM-SHA-256")
	if err != nil {
		t.Fatalf("failed to delete SCRAM-SHA-256: %v", err)
	}

	// 7. Verify only SHA-512 remains
	_, err = kafkaCtl.Execute("describe", "user", username)
	if err != nil {
		t.Fatalf("failed to describe user after partial delete: %v", err)
	}
	output = kafkaCtl.GetStdOut()
	if strings.Contains(output, "SCRAM-SHA-256") {
		t.Fatalf("SCRAM-SHA-256 should be deleted, got: %s", output)
	}
	if !strings.Contains(output, "SCRAM-SHA-512") {
		t.Fatalf("SCRAM-SHA-512 should remain, got: %s", output)
	}

	// 8. Delete remaining mechanism
	_, err = kafkaCtl.Execute("delete", "user", username, "--mechanism", "SCRAM-SHA-512")
	if err != nil {
		t.Fatalf("failed to delete SCRAM-SHA-512: %v", err)
	}

	// 9. Verify user is gone or has no credentials
	_, err = kafkaCtl.Execute("describe", "user", username)
	if err == nil {
		// If no error, check that user has no mechanisms
		output = kafkaCtl.GetStdOut()
		if strings.Contains(output, "SCRAM-") {
			t.Fatalf("user should have no SCRAM credentials after full deletion, got: %s", output)
		}
	}
	// If error occurred, that's also acceptable (user completely gone)
}
