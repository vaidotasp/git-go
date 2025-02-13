package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Entry struct {
	entry_type int
	name       string
	sha        string
}

func readFilesFromSHA1(sha1 string) []Entry {
	// files are zlib compressed
	// dir name is the first two chars
	dir_name := sha1[:2]
	// rest of the chars is the filename
	file_name := sha1[2:]

	// objects always live in the same .git/objects directory
	file_path, err := filepath.Abs(".git/objects/" + dir_name + "/" + file_name)
	if err != nil {
		fmt.Println("Error reading: ", err.Error())
	}

	// open file
	file, err := os.Open(file_path)
	if err != nil {
		log.Fatal(err)
	}

	// zlib decoding
	r, err := zlib.NewReader(file)
	if err != nil {
		fmt.Println("Cannot read the file: ", file_path)
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

	// header := string(file_output.Bytes()[:null_byte_idx]) // tree 99 (type of object and followed by byte size)

	// for the rest of the bytes, we go over each byte, when we find null byte, we split that left side as it contains filetype and filename, then we take the right 20byte SHA and add it to the same entry, increment index by 20 and keep going
	rest_of_object_raw := file_output.Bytes()[null_byte_idx+1:]

	var entries []Entry

	for i := 0; i < len(rest_of_object_raw); {
		v := rest_of_object_raw[i]
		if v == 0 {
			first_part := rest_of_object_raw[:i]
			// check that we have at least 20 bytes to the right
			if i+1+20 > len(rest_of_object_raw) {
				fmt.Println("SHA corruption, not enough bytes left")
				break
			}

			sha_part := rest_of_object_raw[i+1 : i+1+20]

			type_and_name := strings.Split(string(first_part), " ")
			entry_type, err := strconv.Atoi(type_and_name[0])
			if err != nil {
				fmt.Println("Error:", err)
			}

			shaHex := hex.EncodeToString(sha_part)

			entry := Entry{
				entry_type: entry_type,
				name:       type_and_name[1],
				sha:        shaHex,
			}

			entries = append(entries, entry)

			// incrementing by 20 because SHA is 20 length
			i += 1 + 20

			// slicing the main byte array here
			rest_of_object_raw = rest_of_object_raw[i:]

			// reset i to zero because we sliced the main arr above
			i = 0
		} else {
			i++
		}
	}

	return entries
}

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

	case "ls-tree":
		args := os.Args
		if len(args) == 2 {
			fmt.Print("2 args too few")
			os.Exit(1)
		}

		if len(args) == 3 {
			// check that we have 40 length hash
			fmt.Printf(args[2])
			os.Exit(1)
		}

		if len(args) == 4 && args[2] == "--name-only" {
			// check that we have --name-only flag and it is followed by 40 length hash
			sha := args[3]
			s := readFilesFromSHA1(sha)

			var output string

			for _, v := range s {
				file_name := v.name
				output = output + file_name + "\n"
			}

			fmt.Print(output)
		} else {
			sha := args[3]
			s := readFilesFromSHA1(sha)
			var output string

			for _, v := range s {
				// todo:, how to get tree/blob etc?
				output = output + fmt.Sprintf("%d", v.entry_type) + " " + v.name + " " + v.sha + "\n"
			}

			fmt.Print(output)
			// 040000 tree <tree_sha_1>    dir1
			// 040000 tree <tree_sha_2>    dir2
			// 100644 blob <blob_sha_1>    file1
		}
		// tester will check this --name-only, but we can also implement flagless version too

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
