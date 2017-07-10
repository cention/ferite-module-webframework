package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"c3/ferite"
)

const (
	configPublicPath = "/cention/share/ferite/webframework/Public/"
)

// JSProcessString replaces special characters
func JSProcessString(key string) string {
	key = strings.Replace(key, "\\\\", "\\\\\\", -1)
	key = strings.Replace(key, "'", "\\'", -1)
	return key
}

func translationsForJavascript(language string, catalog map[string]interface{}) string {
	output := "// Desired language: " + language + "\n"
	for key, translation := range catalog {
		translation = strings.Replace(translation.(string), "'", "\\'", -1)
		output += "TranslationDictionary['" + JSProcessString(key) + "'] = '" + translation.(string) + "';\n"
	}
	return output
}

func main() {
	if len(os.Args) <= 3 {
		os.Exit(1)
	}
	outputPath := configPublicPath + "Resources/Javascript/Generated"
	webroot := os.Args[1]         // /cention/webroot
	applicationName := os.Args[2] //  Cention
	language := os.Args[3]        //  sv, nb or fi

	translationFile := webroot + "/" + applicationName + ".app/Resources/Strings/" + language + "/translation.strings.fe"
	javascriptFile := outputPath + "/" + applicationName + ".translation." + language + ".js"

	// Open translation mapping source e.g. /cention/webroot/Cention.app/Resources/Strings/[sv|nb|fi]/translation.strings.fe
	ioReader, err := os.Open(translationFile)
	if err != nil {
		log.Printf("Unable to open translation file %s: %s\n", translationFile, err.Error())
		os.Exit(1)
	}
	defer ioReader.Close()

	// Read full content into byte array
	bytesRead, err := ioutil.ReadAll(ioReader)
	if err != nil {
		log.Printf("Unable to load translation for language %s: %s\n", language, err.Error())
		os.Exit(1)
	}
	// As the content is a ferite code of returning an array, thus trim the 'return' keyword and its suffix semi-colon
	// then converts it (ferite code's array) into map[string]interface{}
	content := bytesRead[len("return ") : len(bytesRead)-1]
	translation, err := ferite.DecodeDict(string(content))
	if err != nil {
		log.Printf("Unable to decode translation map: %s\n", err.Error())
		os.Exit(1)
	}
	if err = os.MkdirAll(outputPath, 0755); err != nil {
		log.Printf("Unable to create path %s: %s\n", outputPath, err.Error())
		os.Exit(1)
	}
	// Create the output file e.g. /cention/share/ferite/webframework/Public/Resources/Javascript/Generated/Cention.translation.[sv|nb|fi].js
	ioWriter, err := os.Create(javascriptFile)
	if err != nil {
		log.Printf("Unable to create output file %s: %s\n", javascriptFile, err.Error())
		os.Exit(1)
	}
	defer ioWriter.Close()
	ioWriter.WriteString(translationsForJavascript(language, translation) + "\n")

	log.Println("Created", javascriptFile)
	os.Exit(0)
}
