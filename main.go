package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type commandFuncWithIndex func([]string, int, int) []string
type commandFunc func(string) string

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <input filename> <output filename>")
		return
	}

	fileName := os.Args[1]
	newFileName := os.Args[2]

	if !readWriteFiles(fileName, newFileName) {
		return
	}
}

func readWriteFiles(fileName string, newFileName string) bool {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false
	}

	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println("Error closing file:", err)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return false
	}

	text := fixFile(data)

	newFile, err := os.Create(newFileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return false
	}

	defer func() {
		if err := newFile.Close(); err != nil {
			fmt.Println("Error closing new file:", err)
		}
	}()

	_, err = newFile.Write([]byte(text))
	if err != nil {
		fmt.Println("Error writing to destination file:", err)
		return false
	}
	return true
}

func fixFile(data []byte) string {
	words := strings.Fields(string(data))
	words = convertText(words)
	words = fixIndefiniteArticles(words)
	text := strings.Join(words, " ")
	text = fixPunctuationSpacing(text)
	text = fixSingleQuotes(text)
	return text
}

func convertText(words []string) []string {
	commandsWithIndex := map[string]commandFuncWithIndex{
		"(cap)": capCommand,
		"(up)":  upCommand,
		"(low)": lowCommand,
	}

	commands := map[string]commandFunc{
		"(hex)": hexCommand,
		"(bin)": binCommand,
	}

	var result []string
	for i := 0; i < len(words); i++ {
		word := words[i]
		if strings.HasPrefix(word, "(") && strings.Contains(word, ",") {
			if !strings.HasSuffix(word, ")") && i+1 < len(words) {
				word = word + " " + words[i+1]
				i++
			}
			cmdParts := strings.Split(word, ",")
			cmd := strings.TrimSpace(cmdParts[0])
			if cmdFunc, exists := commandsWithIndex[cmd+")"]; exists && len(cmdParts) == 2 {
				numStr := strings.TrimSpace(cmdParts[1])
				num, err := strconv.Atoi(numStr)
				numStr = strings.TrimRight(numStr, ")")
				if err == nil && len(result) > 0 {
					result = cmdFunc(result, len(result)-1, num)
				}
			}
		} else if cmdFunc, exists := commands[word]; exists {
			if len(result) > 0 {
				result[len(result)-1] = cmdFunc(result[len(result)-1])
			}
		} else {
			result = append(result, word)
		}
	}
	return result
}

func capCommand(result []string, index int, count int) []string {
	for i := 0; i < count && index-i >= 0; i++ {
		result[index-i] = capitalizeWord(result[index-i])
	}
	return result
}

func upCommand(result []string, index int, count int) []string {
	for i := 0; i < count && index-i >= 0; i++ {
		result[index-i] = strings.ToUpper(result[index-i])
	}
	return result
}

func lowCommand(result []string, index int, count int) []string {
	for i := 0; i < count && index-i >= 0; i++ {
		result[index-i] = strings.ToLower(result[index-i])
	}
	return result
}

func hexCommand(result string) string {
	decimalValue, err := strconv.ParseInt(result, 16, 64)
	if err != nil {
		fmt.Println("Error converting hex to dec")
		return result
	}
	return strconv.FormatInt(decimalValue, 10)
}

func binCommand(result string) string {
	decimalValue, err := strconv.ParseInt(result, 2, 64)
	if err != nil {
		fmt.Println("Error converting bin to dec")
		return result
	}
	return strconv.FormatInt(decimalValue, 10)
}

func capitalizeWord(word string) string {
	if len(word) == 0 {
		return word
	}
	return string(unicode.ToUpper(rune(word[0]))) + strings.ToLower(word[1:])
}

func fixPunctuationSpacing(text string) string {
	var result strings.Builder
	runes := []rune(text)
	length := len(runes)
	targetPunctuations := ".!,?:;"

	for i := 0; i < length; i++ {
		current := runes[i]

		if current == ' ' && i+1 < length && strings.ContainsRune(targetPunctuations, runes[i+1]) {
			// Skip the space
			continue
		}

		result.WriteRune(current)

		if strings.ContainsRune(targetPunctuations, current) {
			if i+1 < length {
				next := runes[i+1]
				if next != ' ' && !strings.ContainsRune(targetPunctuations, next) {
					result.WriteRune(' ')
				}
			}
		}
	}
	return result.String()
}

func fixSingleQuotes(text string) string {
	var result strings.Builder
	runes := []rune(text)
	length := len(runes)
	inSingleQuote := false

	for i := 0; i < length; i++ {
		current := runes[i]

		if current == '\'' {
			inSingleQuote = !inSingleQuote
			result.WriteRune(current)
			continue
		}

		if inSingleQuote {
			if (i > 0 && runes[i-1] == '\'' && current == ' ') || (i+1 < length && runes[i+1] == '\'' && current == ' ') {
				continue
			}
		}

		result.WriteRune(current)
	}
	return result.String()
}

func fixIndefiniteArticles(words []string) []string {
	vowelsAndH := "aeiouhAEIOUH"
	for i := 0; i < len(words)-1; i++ {
		currentWord := words[i]
		nextWord := words[i+1]

		currentWordStripped := strings.TrimRightFunc(currentWord, func(r rune) bool {
			return unicode.IsPunct(r)
		})

		nextWordStripped := strings.TrimLeftFunc(nextWord, func(r rune) bool {
			return unicode.IsPunct(r)
		})

		if (currentWordStripped == "a" || currentWordStripped == "A") && len(nextWordStripped) > 0 {
			firstLetter := rune(nextWordStripped[0])
			if strings.ContainsRune(vowelsAndH, unicode.ToLower(firstLetter)) {
				replacement := "an"
				if currentWordStripped == "A" {
					replacement = "An"
				}
				trailingPunct := currentWord[len(currentWordStripped):]
				words[i] = replacement + trailingPunct
			}
		}
	}
	return words
}
