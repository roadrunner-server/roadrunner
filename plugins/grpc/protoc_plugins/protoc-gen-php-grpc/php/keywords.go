// MIT License
//
// Copyright (c) 2018 SpiralScout
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package php

import (
	"bytes"
	"strings"
	"unicode"
)

// @see https://github.com/protocolbuffers/protobuf/blob/master/php/ext/google/protobuf/protobuf.c#L168
var reservedKeywords = []string{
	"abstract", "and", "array", "as", "break",
	"callable", "case", "catch", "class", "clone",
	"const", "continue", "declare", "default", "die",
	"do", "echo", "else", "elseif", "empty",
	"enddeclare", "endfor", "endforeach", "endif", "endswitch",
	"endwhile", "eval", "exit", "extends", "final",
	"for", "foreach", "function", "global", "goto",
	"if", "implements", "include", "include_once", "instanceof",
	"insteadof", "interface", "isset", "list", "namespace",
	"new", "or", "print", "private", "protected",
	"public", "require", "require_once", "return", "static",
	"switch", "throw", "trait", "try", "unset",
	"use", "var", "while", "xor", "int",
	"float", "bool", "string", "true", "false",
	"null", "void", "iterable",
}

// Check if given name/keyword is reserved by php.
func isReserved(name string) bool {
	name = strings.ToLower(name)
	for _, k := range reservedKeywords {
		if name == k {
			return true
		}
	}

	return false
}

// generate php namespace or path
func namespace(pkg *string, sep string) string {
	if pkg == nil {
		return ""
	}

	result := bytes.NewBuffer(nil)
	for _, p := range strings.Split(*pkg, ".") {
		result.WriteString(identifier(p, ""))
		result.WriteString(sep)
	}

	return strings.Trim(result.String(), sep)
}

// create php identifier for class or message
func identifier(name string, suffix string) string {
	name = Camelize(name)
	if suffix != "" {
		return name + Camelize(suffix)
	}

	return name
}

func resolveReserved(identifier string, pkg string) string {
	if isReserved(strings.ToLower(identifier)) {
		if pkg == ".google.protobuf" {
			return "GPB" + identifier
		}
		return "PB" + identifier
	}

	return identifier
}

// Camelize "dino_party" -> "DinoParty"
func Camelize(word string) string {
	words := splitAtCaseChangeWithTitlecase(word)
	return strings.Join(words, "")
}

func splitAtCaseChangeWithTitlecase(s string) []string {
	words := make([]string, 0)
	word := make([]rune, 0)
	for _, c := range s {
		spacer := isSpacerChar(c)
		if len(word) > 0 {
			if unicode.IsUpper(c) || spacer {
				words = append(words, string(word))
				word = make([]rune, 0)
			}
		}
		if !spacer {
			if len(word) > 0 {
				word = append(word, unicode.ToLower(c))
			} else {
				word = append(word, unicode.ToUpper(c))
			}
		}
	}
	words = append(words, string(word))
	return words
}

func isSpacerChar(c rune) bool {
	switch {
	case c == rune("_"[0]):
		return true
	case c == rune(" "[0]):
		return true
	case c == rune(":"[0]):
		return true
	case c == rune("-"[0]):
		return true
	}
	return false
}
