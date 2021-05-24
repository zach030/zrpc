package codec

import (
	"fmt"
	"strings"
	"testing"
)

func TestGobCodec_Write(t *testing.T) {
	res := checkInclusion("ab", "eidboaoo")
	fmt.Println(res)
}

type ListNode struct {
	Val  int
	Next *ListNode
}

func reverseList(head *ListNode) *ListNode {
	var prev *ListNode
	curr := head
	for curr != nil {
		next := curr.Next
		curr.Next = prev
		prev = curr
		curr = next
	}
	return prev
}

func twoSum(nums []int, target int) []int {
	res := make([]int, 0)
	kv := make(map[int]int, 1)
	for i := 0; i < len(nums); i++ {
		if _, ok := kv[nums[i]]; ok {
			res = append(res, i)
			res = append(res, kv[nums[i]])
			return res
		}
		kv[target-nums[i]] = i
	}
	return nil
}

func findDisappearedNumbers(nums []int) []int {
	size := len(nums)
	filter := make(map[int]bool, 0)
	res := make([]int, 0)
	for i := 1; i < size; i++ {
		filter[i] = false
	}
	for _, num := range nums {
		filter[num] = true
	}
	for k, v := range filter {
		if v == false {
			res = append(res, k)
		}
	}
	return res
}

func maxProfit(prices []int) int {
	ans := 0
	cur := 0
	for i := 1; i < len(prices); i++ {
		if prices[i] > prices[cur] {
			ans = getMax(ans, prices[i]-prices[cur])
		} else {
			cur++
		}
	}
	return ans
}

func getMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func merge(nums1 []int, m int, nums2 []int, n int) {
	p := m + n - 1
	m--
	n--
	for m >= 0 && n >= 0 {
		if nums1[m] > nums2[n] {
			nums1[p] = nums1[m]
			m--
		} else {
			nums1[p] = nums2[n]
			n--
		}
		p--
	}
	for n >= 0 {
		nums1[p] = nums2[n]
		p--
		n--
	}
	return
}

func missingNumber(nums []int) int {
	kv := make(map[int]bool, 0)
	for i := 0; i <= len(nums); i++ {
		kv[i] = true
	}

	for _, num := range nums {
		if _, ok := kv[num]; !ok {
			return num
		}
	}
	return 0
}

func checkInclusion(s1 string, s2 string) bool {
	i := 0
	for i < len(s2) {
		flag := true
		if contain(s1, s2[i]) {
			// 接下来len(s1)-1个都应该contain
			j := len(s1) - 1
			index := i + 1

			for j > 0 {
				if !contain(s1, s2[index]) {
					flag = false
					break
				}
				index++
				j--
			}
		}
		if flag {
			return true
		}
		i++
	}
	return false
}

func contain(nums string, s byte) bool {
	return strings.Contains(nums, string(s))
}
