package cube_test

import (
	"testing"

	"cube-auth/internal/cube"
)

// TestColorCountImbalanceIsRejected guards the former counts[color] < 7
// threshold, which accepted eight/ten sticker counts.
func TestColorCountImbalanceIsRejected(t *testing.T) {
	faces := solved()
	// Overwrite one U sticker with the D-center color: U's color now occurs
	// eight times and D's color ten times (a simple swap would keep both at
	// nine and slip past the count check legitimately).
	faces["U"][1] = faces["D"][4]

	report := cubeValidate(t, faces)
	if report.Legal || report.Violation.Code != "colors" {
		t.Fatalf("got legal=%v violation=%#v, want colors violation", report.Legal, report.Violation)
	}
	found8, found10 := false, false
	for _, issue := range report.Violation.Colors {
		if issue.Kind == "wrong_count" && issue.Count == 8 {
			found8 = true
		}
		if issue.Kind == "wrong_count" && issue.Count == 10 {
			found10 = true
		}
	}
	if !found8 || !found10 {
		t.Fatalf("color issues = %#v, want wrong_count entries for 8 and 10", report.Violation.Colors)
	}
}

// TestLegalReportRoundTripsStickers verifies that the cubie representation
// returned for legal states reconstructs the exact input sticker colors.
func TestLegalReportRoundTripsStickers(t *testing.T) {
	moveSets := [][]byte{
		{},
		{'U'}, {'R'}, {'F'}, {'D'}, {'L'}, {'B'},
		{'U', 'R', 'F', 'D', 'L', 'B'},
		{'U', 'U', 'R', 'R', 'R', 'F', 'L', 'L', 'D', 'B', 'B', 'B'},
	}
	for _, moves := range moveSets {
		faces := solved()
		for _, m := range moves {
			faces = turn(faces, m)
		}
		report := cubeValidate(t, faces)
		if !report.Legal {
			t.Fatalf("moves %q rejected: %#v", string(moves), report.Violation)
		}
		roundTripAndCompare(t, faces, report.Cube, string(moves))
	}
}

// TestThreeEdgeCycleIsLegal guards that only differing parity is flagged.
func TestThreeEdgeCycleIsLegal(t *testing.T) {
	// A three-cycle of edges is even; corners are untouched.
	faces := cycleEdges(solved(),
		[2]sticker{{"U", 5}, {"R", 1}}, // UR
		[2]sticker{{"U", 7}, {"F", 1}}, // UF
		[2]sticker{{"U", 3}, {"L", 1}}, // UL
	)
	report := cubeValidate(t, faces)
	if !report.Legal {
		t.Fatalf("even three-cycle of edges rejected: %#v", report.Violation)
	}
}

// TestSingleCornerSwapHasParityViolation verifies the parity check is not
// gated on both permutations being odd (corners odd, edges even must fail).
func TestSingleCornerSwapHasParityViolation(t *testing.T) {
	faces := solved()
	swapStickerSlots(faces,
		[3]sticker{{"U", 8}, {"R", 0}, {"F", 2}}, // URF
		[3]sticker{{"U", 6}, {"F", 0}, {"L", 2}}, // UFL
	)
	report := cubeValidate(t, faces)
	if report.Legal {
		t.Fatal("single transposition of two corners accepted")
	}
	if report.Violation.Code != "permutation_parity" {
		t.Fatalf("violation = %q, want permutation_parity", report.Violation.Code)
	}
}

func cubeValidate(t *testing.T, faces cube.Faces) *cube.Report {
	t.Helper()
	return cube.Validate(faces)
}

// roundTripAndCompare rebuilds every non-center sticker from the cubie
// representation and checks it against the original facelet colors.
func roundTripAndCompare(t *testing.T, faces cube.Faces, c *cube.CubieCube, label string) {
	t.Helper()
	centers := map[string]string{}
	for _, name := range []string{"U", "R", "F", "D", "L", "B"} {
		centers[name] = faces[name][4]
	}
	rebuilt := cube.Faces{}
	for _, name := range []string{"U", "R", "F", "D", "L", "B"} {
		rebuilt[name] = make([]string, 9)
		rebuilt[name][4] = centers[name]
	}
	put := func(refs []struct {
		face  string
		index int
	}, colors []string) {
		for i, ref := range refs {
			rebuilt[ref.face][ref.index] = colors[i]
		}
	}

	cornerSlots := [8][3]sticker{
		{{"U", 8}, {"R", 0}, {"F", 2}},
		{{"U", 6}, {"F", 0}, {"L", 2}},
		{{"U", 0}, {"L", 0}, {"B", 2}},
		{{"U", 2}, {"B", 0}, {"R", 2}},
		{{"D", 2}, {"F", 8}, {"R", 6}},
		{{"D", 0}, {"L", 8}, {"F", 6}},
		{{"D", 6}, {"B", 8}, {"L", 6}},
		{{"D", 8}, {"R", 8}, {"B", 6}},
	}
	edgeSlots := [12][2]sticker{
		{{"U", 5}, {"R", 1}},
		{{"U", 7}, {"F", 1}},
		{{"U", 3}, {"L", 1}},
		{{"U", 1}, {"B", 1}},
		{{"D", 5}, {"R", 7}},
		{{"D", 1}, {"F", 7}},
		{{"D", 3}, {"L", 7}},
		{{"D", 7}, {"B", 7}},
		{{"F", 5}, {"R", 3}},
		{{"F", 3}, {"L", 5}},
		{{"B", 5}, {"L", 3}},
		{{"B", 3}, {"R", 5}},
	}

	for pos := 0; pos < 8; pos++ {
		piece := c.CornerPositions[pos]
		ori := c.CornerOrientations[pos]
		colors := [3]string{}
		for k := 0; k < 3; k++ {
			home := cornerSlots[piece][(k-ori+3)%3]
			colors[k] = centers[home.face]
		}
		refs := make([]struct {
			face  string
			index int
		}, 3)
		for k := range refs {
			refs[k] = struct {
				face  string
				index int
			}{cornerSlots[pos][k].face, cornerSlots[pos][k].index}
		}
		put(refs, colors[:])
	}
	for pos := 0; pos < 12; pos++ {
		piece := c.EdgePositions[pos]
		ori := c.EdgeOrientations[pos]
		colors := [2]string{
			centers[edgeSlots[piece][ori].face],
			centers[edgeSlots[piece][(ori+1)%2].face],
		}
		for k := 0; k < 2; k++ {
			ref := edgeSlots[pos][k]
			rebuilt[ref.face][ref.index] = colors[k]
		}
	}

	for _, name := range []string{"U", "R", "F", "D", "L", "B"} {
		for i := 0; i < 9; i++ {
			if rebuilt[name][i] != faces[name][i] {
				t.Errorf("moves %q: rebuilt %s[%d]=%q, input has %q",
					label, name, i, rebuilt[name][i], faces[name][i])
			}
		}
	}
}

// cycleEdges rotates the pieces at positions a<-c, b<-a, c<-b, keeping each
// edge's two stickers in slot order (no flips).
func cycleEdges(faces cube.Faces, a, b, c [2]sticker) cube.Faces {
	av := [2]string{faces[a[0].face][a[0].index], faces[a[1].face][a[1].index]}
	bv := [2]string{faces[b[0].face][b[0].index], faces[b[1].face][b[1].index]}
	cv := [2]string{faces[c[0].face][c[0].index], faces[c[1].face][c[1].index]}
	for k := 0; k < 2; k++ {
		faces[b[k].face][b[k].index] = av[k]
		faces[c[k].face][c[k].index] = bv[k]
		faces[a[k].face][a[k].index] = cv[k]
	}
	return faces
}

// swapStickerSlots exchanges all stickers of two corner slot triples in
// slot order, transposing the two corners without twisting them.
func swapStickerSlots(faces cube.Faces, a, b [3]sticker) {
	for k := 0; k < 3; k++ {
		faces[a[k].face][a[k].index], faces[b[k].face][b[k].index] =
			faces[b[k].face][b[k].index], faces[a[k].face][a[k].index]
	}
}
