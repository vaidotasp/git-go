package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func generateSHA1(uncompressed_content []byte) string {
	header := []byte("blob ")
	sizeStr := strconv.Itoa(len(uncompressed_content))
	header = append(header, []byte(sizeStr)...)
	header = append(header, 0)
	header = append(header, uncompressed_content...)
	hash := sha1.Sum(header)
	return fmt.Sprintf("%x", hash)
}

// Usage: your_program.sh <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintf(os.Stderr, "Logs from your program will appear here!\n")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		// Uncomment this block to pass the first stage!
		//
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}
		//
		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")

	case "cat-file":
		args := os.Args

		// no out of bounds errors please
		if len(args) > 3 {

			path_cmd := args[2]
			blob_sha := args[3]

			// check that we have -p flag and it is followed by 40 length blob sha
			if path_cmd != "-p" && len(blob_sha) != 40 {
				fmt.Fprintf(os.Stderr, "Missing -p flag or incorrect Blob SHA")
				os.Exit(1)
			}

			dir_name := blob_sha[:2]
			file_name := blob_sha[2:]

			file_path, err := filepath.Abs(".git/objects/" + dir_name + "/" + file_name)
			if err != nil {
				fmt.Println("Error reading: ", err.Error())
			}

			file, err := os.Open(file_path) // For read access.
			if err != nil {
				log.Fatal(err)
			}

			r, err := zlib.NewReader(file)
			if err != nil {
				fmt.Println("Cant read file")
			}
			defer r.Close()

			var file_output bytes.Buffer
			_, err = io.Copy(&file_output, r)
			if err != nil {
				log.Fatal(err)
			}

			// iterate over bytes and find null byte, after null byte we have the output of the file that we care about
			var null_byte_idx int
			for k, v := range file_output.Bytes() {
				if v == 0 {
					null_byte_idx = k
					break
				}
			}
			byte_output := file_output.Bytes()[null_byte_idx+1:]
			decompressedString := string(byte_output[:])
			fmt.Print(decompressedString)
		}
	case "hash-object":
		args := os.Args
		path_cmd := args[2]

		// check that we have -p flag and it is followed by 40 length blob sha
		if path_cmd != "-w" {
			fmt.Fprintf(os.Stderr, "Missing -w flag or incorrect Blob SHA")
			os.Exit(1)
		}

		input_file := args[3]
		data, err := os.ReadFile(input_file) // For read access.
		if err != nil {
			log.Fatal(err)
		}

		var buf bytes.Buffer
		w := zlib.NewWriter(&buf)
		w.Write([]byte(data))
		w.Close()

		r, err := zlib.NewReader(&buf)
		if err != nil {
		}
		defer r.Close()

		var uncompressed bytes.Buffer
		_, err = io.Copy(&uncompressed, r)
		if err != nil {
			log.Fatal(err)
		}

		hash := generateSHA1(data)
		dir_name := hash[:2]
		file_name := hash[2:]

		dir_path, err := filepath.Abs(".git/objects/")
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
		}

		if err := os.MkdirAll(dir_path+"/"+dir_name, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
		}

		//writing file
		file_path := dir_path + "/" + dir_name + "/" + file_name
		sizeStr := strconv.Itoa(len(data))
		output := []byte("blob ")
		uncompressed_file_content := uncompressed.String()
		output = append(output, []byte(sizeStr)...)
		output = append(output, 0)
		output = append(output, uncompressed_file_content...)

		var b bytes.Buffer
		w = zlib.NewWriter(&b)
		w.Write([]byte(output))
		w.Close()

		err = os.WriteFile(file_path, b.Bytes(), 0644)
		if err != nil {
			fmt.Println("Error writing: ", err.Error())
		}

		fmt.Print(hash)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
