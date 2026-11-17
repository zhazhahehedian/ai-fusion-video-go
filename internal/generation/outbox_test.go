package generation

import (
	"errors"
	"strings"
	"testing"
)

func TestOutboxErrorTextKeepsPrefix(t *testing.T) {
	msg := outboxErrorText(errors.New("连接中断"))
	if !strings.HasPrefix(msg, "事件投递失败：") || !strings.Contains(msg, "连接中断") {
		t.Fatalf("message = %q", msg)
	}
	long := outboxErrorText(errors.New(strings.Repeat("错", 800)))
	if len([]rune(long)) != 500 {
		t.Fatalf("truncated len = %d", len([]rune(long)))
	}
}

func TestDeliveryFailureUnwraps(t *testing.T) {
	cause := errors.New("未登记")
	err := &deliveryFailure{id: 7, err: cause}
	var failed *deliveryFailure
	if !errors.As(err, &failed) || failed.id != 7 || !errors.Is(err, cause) {
		t.Fatalf("unwrap = %v", err)
	}
}
