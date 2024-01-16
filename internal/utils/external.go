package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

func RunThirdPartyClient[T any](obj *[]T, args ...string) error {
	cmd := exec.Command("sh", args...)
	cmd.Stderr = os.Stderr // or any other io.Writer
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed at running os command: %w", err)
	}
	if obj != nil {
		err = json.Unmarshal(out, &obj)
		if err != nil {
			fmt.Println(err.Error())
			return fmt.Errorf("failed at unmarshalling response: %w", err)
		}
	}
	return nil
}
