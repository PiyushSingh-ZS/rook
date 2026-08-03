package main

import "testing"

func gnode(id, typ string, deps ...graphDep) *graphNode {
	return &graphNode{ID: id, Name: id, Type: typ, DependsOn: deps, Status: "pending", Session: id}
}
func dep(node, on string) graphDep { return graphDep{Node: node, On: on} }

// complete marks a node terminal with a result, as the runtime would.
func complete(n *graphNode, result string) {
	n.Result = result
	if result == "fail" {
		n.Status = "failed"
	} else {
		n.Status = "done"
	}
}
func statusOfNode(g *taskGraph, id string) string { return nodeByID(g, id).Status }
func launchedIDs(ns []*graphNode) []string {
	out := []string{}
	for _, n := range ns {
		out = append(out, n.ID)
	}
	return out
}
func onlyID(t *testing.T, ns []*graphNode, want string) {
	t.Helper()
	if len(ns) != 1 || ns[0].ID != want {
		t.Fatalf("expected launch [%s], got %v", want, launchedIDs(ns))
	}
}

func TestGraphLinearOneAtATime(t *testing.T) {
	g := &taskGraph{Nodes: []*graphNode{
		gnode("a", "agent"),
		gnode("b", "agent", dep("a", "pass")),
		gnode("c", "agent", dep("b", "pass")),
	}}
	l, _ := scheduleGraph(g)
	onlyID(t, l, "a") // only a runnable; single agent at a time
	complete(nodeByID(g, "a"), "pass")
	l, _ = scheduleGraph(g)
	onlyID(t, l, "b")
	complete(nodeByID(g, "b"), "pass")
	l, _ = scheduleGraph(g)
	onlyID(t, l, "c")
	complete(nodeByID(g, "c"), "pass")
	scheduleGraph(g)
	if !g.Done {
		t.Fatalf("graph should be done")
	}
}

func TestGraphConditionalEdges(t *testing.T) {
	build := func() *taskGraph {
		return &taskGraph{Nodes: []*graphNode{
			gnode("a", "agent"),
			gnode("ok", "agent", dep("a", "pass")),
			gnode("bad", "agent", dep("a", "fail")),
		}}
	}
	// a passes → ok runs, bad skipped
	g := build()
	scheduleGraph(g)
	complete(nodeByID(g, "a"), "pass")
	l, _ := scheduleGraph(g)
	onlyID(t, l, "ok")
	if statusOfNode(g, "bad") != "skipped" {
		t.Fatalf("bad should be skipped when a passed, got %s", statusOfNode(g, "bad"))
	}
	// a fails → bad runs, ok skipped
	g = build()
	scheduleGraph(g)
	complete(nodeByID(g, "a"), "fail")
	l, _ = scheduleGraph(g)
	onlyID(t, l, "bad")
	if statusOfNode(g, "ok") != "skipped" {
		t.Fatalf("ok should be skipped when a failed, got %s", statusOfNode(g, "ok"))
	}
}

func TestGraphApprovalInterrupt(t *testing.T) {
	g := &taskGraph{Nodes: []*graphNode{
		gnode("a", "agent"),
		gnode("gate", "approval", dep("a", "pass")),
		gnode("b", "agent", dep("gate", "pass")),
	}}
	scheduleGraph(g)
	complete(nodeByID(g, "a"), "pass")
	l, aw := scheduleGraph(g)
	if len(l) != 0 {
		t.Fatalf("no agent should run while awaiting approval, got %v", launchedIDs(l))
	}
	if len(aw) != 1 || aw[0].ID != "gate" || statusOfNode(g, "gate") != "awaiting" {
		t.Fatalf("gate should be awaiting approval, got %v", launchedIDs(aw))
	}
	// approve
	complete(nodeByID(g, "gate"), "pass")
	l, _ = scheduleGraph(g)
	onlyID(t, l, "b")
}

func TestGraphDiamond(t *testing.T) {
	g := &taskGraph{Nodes: []*graphNode{
		gnode("a", "agent"),
		gnode("b", "agent", dep("a", "pass")),
		gnode("c", "agent", dep("a", "pass")),
		gnode("d", "agent", dep("b", "pass"), dep("c", "pass")),
	}}
	scheduleGraph(g)
	complete(nodeByID(g, "a"), "pass")
	l, _ := scheduleGraph(g) // b or c (single agent) — b comes first in order
	onlyID(t, l, "b")
	complete(nodeByID(g, "b"), "pass")
	l, _ = scheduleGraph(g)
	onlyID(t, l, "c")
	complete(nodeByID(g, "c"), "pass")
	l, _ = scheduleGraph(g)
	onlyID(t, l, "d") // d only after BOTH b and c
}

func TestGraphSkipPropagation(t *testing.T) {
	g := &taskGraph{Nodes: []*graphNode{
		gnode("a", "agent"),
		gnode("b", "agent", dep("a", "pass")),
		gnode("c", "agent", dep("b", "pass")),
	}}
	scheduleGraph(g)
	complete(nodeByID(g, "a"), "fail") // a fails → b needs pass → skip → c needs b pass → skip
	scheduleGraph(g)
	if statusOfNode(g, "b") != "skipped" || statusOfNode(g, "c") != "skipped" {
		t.Fatalf("skip should propagate: b=%s c=%s", statusOfNode(g, "b"), statusOfNode(g, "c"))
	}
	if !g.Done {
		t.Fatalf("graph should be done after all nodes terminal")
	}
}
