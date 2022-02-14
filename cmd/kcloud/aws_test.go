package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAWSProfileRegexp(t *testing.T) {
	assert.False(t, awsProfileRegex.MatchString("foobar"))
	matches := awsProfileRegex.FindStringSubmatch("foobar")
	assert.Equal(t, 0, len(matches))

	assert.True(t, awsProfileRegex.MatchString("  [ foobar ] "))
	matches = awsProfileRegex.FindStringSubmatch("  [  foo bar ] ")
	assert.Equal(t, 2, len(matches))
	assert.Equal(t, "foo bar", strings.TrimSpace(matches[1]))
}

func TestAWSRegionRegexp(t *testing.T) {
	assert.False(t, awsRegionRegex.MatchString("foobar"))
	matches := awsRegionRegex.FindStringSubmatch("foobar")
	assert.Equal(t, 0, len(matches))

	assert.True(t, awsRegionRegex.MatchString(" region = foobar "))
	matches = awsRegionRegex.FindStringSubmatch(" region = foobar ")
	assert.Equal(t, 2, len(matches))
	assert.Equal(t, "foobar", strings.TrimSpace(matches[1]))
}
