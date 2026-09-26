package task

import (
	"reflect"
	"testing"
)

func TestExpandHostPattern_NoBracket(t *testing.T) {
	got, err := ExpandHostPattern("web1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"web1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpandHostPattern_NumericRange(t *testing.T) {
	got, err := ExpandHostPattern("web[1-3]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"web1", "web2", "web3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpandHostPattern_LeadingZero(t *testing.T) {
	got, err := ExpandHostPattern("web[01-03]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"web01", "web02", "web03"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpandHostPattern_LeadingZeroWide(t *testing.T) {
	got, err := ExpandHostPattern("web[01-10]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{
		"web01", "web02", "web03", "web04", "web05",
		"web06", "web07", "web08", "web09", "web10",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpandHostPattern_Enum(t *testing.T) {
	got, err := ExpandHostPattern("web[a,b,c]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"weba", "webb", "webc"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpandHostPattern_Mixed(t *testing.T) {
	got, err := ExpandHostPattern("web[1-2,5,7-8]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"web1", "web2", "web5", "web7", "web8"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpandHostPattern_MultiBracket(t *testing.T) {
	got, err := ExpandHostPattern("web[1-2].dc[beijing,wuhan].com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{
		"web1.dcbeijing.com",
		"web1.dcwuhan.com",
		"web2.dcbeijing.com",
		"web2.dcwuhan.com",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpandHostPattern_IPRange(t *testing.T) {
	got, err := ExpandHostPattern("10.0.0.[1-3]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpandHostPattern_CharRange(t *testing.T) {
	got, err := ExpandHostPattern("node[a-c]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"nodea", "nodeb", "nodec"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpandHostPattern_GroupPrefix(t *testing.T) {
	got, err := ExpandHostPattern("prod/web[1-3]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"prod/web1", "prod/web2", "prod/web3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExpandHostPattern_InvalidRange(t *testing.T) {
	_, err := ExpandHostPattern("web[5-1]")
	if err == nil {
		t.Fatal("expected error for invalid range, got nil")
	}
}

func TestExpandHostPattern_UnclosedBracket(t *testing.T) {
	_, err := ExpandHostPattern("web[1-3")
	if err == nil {
		t.Fatal("expected error for unclosed bracket, got nil")
	}
}

func TestExpandHostPattern_Empty(t *testing.T) {
	got, err := ExpandHostPattern("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestExpandHostPattern_SingleItem(t *testing.T) {
	got, err := ExpandHostPattern("web[5]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"web5"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
