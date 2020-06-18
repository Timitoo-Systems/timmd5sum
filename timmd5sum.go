package main

import (
    "crypto/md5"
    "flag"
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"
    "strings"
    "text/template"
)

type TemplateData struct {
    InFilePath string // Input file to hash
    Name   string  // Arbitrary name
    Config string  // Name of the compilation configuration
    Md5Sum string  // md5 sum of input file
    OutDir string // optional
    OutFilePath string // the file containing the md5
}

func (td *TemplateData) upper() {
   td.Name = strings.ToUpper(td.Name)
   td.Config = strings.ToUpper(td.Config)
}

func renderTemplate(outfile * os.File, templData *TemplateData) {

    templData.upper()

    var templ string =
`
// {{.InFilePath}}
#define {{.Name}}_MD5_{{.Config}} "{{.Md5Sum}}" 
`

    t, err := template.New("c_header").Parse(templ)
    if err != nil {
        panic(err)
    }

    err = t.Execute(outfile, templData)
    if err != nil {
        panic(err)
    }

}

func md5sum(td * TemplateData) bool{
    var rc bool = true
    var in string = td.InFilePath

    f, err := os.Open(in)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    h := md5.New()
    if _, err := io.Copy(h, f); err != nil {
        log.Fatal(err)

    }

    td.Md5Sum = fmt.Sprintf("%x", h.Sum(nil))
    return rc
}

func myUsage() {
    fmt.Printf("Usage: %s [OPTIONS] file ...\n", "timmd5sum")
    flag.PrintDefaults()
}

func parseFlags(td * TemplateData) bool {
    var rc bool = true

    flag.Usage = myUsage

    namePtr := flag.String("n", "", "Arbitrary name for the sum")
    configPtr := flag.String("c", "", "Configuration information (Debug/Release/RelWithDbgInfo)")
    outdirPtr := flag.String("o", "", "Output directory")

    flag.Parse()

    var inFiles []string = flag.Args()
    if len(inFiles) == 0 {
        flag.Usage()
        os.Exit(1)
    } else {
        td.InFilePath = inFiles[0]
    }

    var outfilename string
    outfilename = strings.ToLower(*namePtr)
    outfilename += "_"
    outfilename += strings.ToLower(*configPtr)
    outfilename += "_md5.h"

    if *outdirPtr == "" {
        td.OutFilePath = outfilename
    } else {
        td.OutDir = *outdirPtr
        td.OutFilePath = filepath.Join(*outdirPtr,outfilename,)
    }

    td.Name = *namePtr
    td.Config = *configPtr

    return rc
}


func main() {

    td := TemplateData{}

    parseFlags(&td)

    if md5sum(&td) == false {
        return
    }

    // check dir path
    if _, err := os.Stat(td.OutDir); os.IsNotExist(err) {
        // path does not exist
        _ = os.MkdirAll(td.OutDir, 777)
    }

    oFile, err := os.Create(td.OutFilePath) // rw create trunc
    if err != nil {
        log.Fatalf("Failed to open %s", td.OutFilePath)
    }

    defer oFile.Close()

    renderTemplate(oFile, &td)

    os.Exit(0)
}
