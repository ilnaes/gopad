package internal

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// create a diff between two strings
// returns list of operations to change
// s1 into s2
func diff(s1, s2 []byte) []Op {
	dp := make([][]int, len(s1)+1)
	dp[0] = make([]int, len(s2)+1)

	// DP to calculate diff
	// TODO: maybe make top-down
	for j := 0; j < len(s2)+1; j++ {
		dp[0][j] = j
	}

	for i := 1; i < len(s1)+1; i++ {
		dp[i] = make([]int, len(s2)+1)
		dp[i][0] = i

		for j := 1; j < len(s2)+1; j++ {
			dp[i][j] = min(dp[i][j-1], dp[i-1][j]) + 1

			if s1[i-1] == s2[j-1] && dp[i-1][j-1] < dp[i][j] {
				dp[i][j] = dp[i-1][j-1]
			}
		}
	}

	i := len(s1)
	j := len(s2)

	res := []Op{}

	// collect diff into slice
	for i > 0 || j > 0 {
		if i == 0 {
			res = append(res, Op{Add: true, Loc: i, Ch: s2[j-1]})
			j--
		} else if j == 0 {
			res = append(res, Op{Add: false, Loc: i - 1, Ch: s1[i-1]})
			i--
		} else {
			if s1[i-1] == s2[j-1] && dp[i][j] == dp[i-1][j-1] {
				i--
				j--
			} else {
				if dp[i][j] == dp[i][j-1]+1 {
					// Add s2[j-1]
					res = append(res, Op{Add: true, Loc: i, Ch: s2[j-1]})
					j--
				} else {
					// resete s1[i-1]
					res = append(res, Op{Add: false, Loc: i - 1, Ch: s1[i-1]})
					i--
				}
			}
		}
	}

	i = 0
	j = len(res) - 1

	// reverse order
	for i < j {
		res[i], res[j] = res[j], res[i]
		i++
		j--
	}

	return res
}

// applies a set of operations (in increasing Location
// order) to a byte slice
func Apply(s []byte, ops []Op) []byte {
	res := []byte{}

	i := 0

	for _, op := range ops {
		res = append(res, s[i:op.Loc]...)
		i = op.Loc

		if op.Add {
			res = append(res, op.Ch)
		} else {
			i++
		}
	}
	if i < len(s) {
		res = append(res, s[i:]...)
	}

	return res
}

// operational transform o2 in the event that o1 gets
// applied first; both are in increasing Loc order
func Xform(o1, o2 []Op) []Op {
	res := []Op{}

	i := 0
	j := 0
	delta := 0

	for j < len(o2) {
		if i == len(o1) || o1[i].Loc > o2[j].Loc {
			res = append(res, o2[j])
			res[len(res)-1].Loc += delta
			j++
		} else if o1[i].Loc == o2[j].Loc {
			if !o1[i].Add && !o2[j].Add {
				// two deletes so skip
				j++
				i++
				delta--
			} else {
				// do Add first
				if o1[i].Add {
					delta++
					i++
				} else {
					res = append(res, o2[j])
					res[len(res)-1].Loc += delta
					j++
				}
			}
		} else {
			if o1[i].Add {
				delta++
			} else {
				delta--
			}
			i++
		}
	}

	return res
}
