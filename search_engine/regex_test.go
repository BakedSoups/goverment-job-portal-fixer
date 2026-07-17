package search_engine

import "testing"

func TestMatcherRequiresITContextForInfrastructure(t *testing.T) {
	matcher := NewMatcher()

	physical := matcher.Match("Lead waterfront infrastructure development, construction coordination, and climate resilience projects.")
	if physical["cloud_infrastructure"] != 0 {
		t.Fatalf("physical infrastructure matched cloud_infrastructure: %#v", physical)
	}

	technical := matcher.Match("Manage AWS cloud infrastructure with Terraform, Kubernetes, and disaster recovery controls.")
	if technical["cloud_infrastructure"] == 0 {
		t.Fatalf("technical infrastructure did not match cloud_infrastructure: %#v", technical)
	}
}

func TestMatcherRequiresEnterpriseContextForArchitecture(t *testing.T) {
	matcher := NewMatcher()

	building := matcher.Match("Review historic architecture and building preservation requirements.")
	if building["enterprise_architecture"] != 0 {
		t.Fatalf("building architecture matched enterprise_architecture: %#v", building)
	}

	technical := matcher.Match("Define enterprise architecture and systems architecture standards.")
	if technical["enterprise_architecture"] == 0 {
		t.Fatalf("technical architecture did not match enterprise_architecture: %#v", technical)
	}
}
