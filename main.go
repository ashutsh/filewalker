package main

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"


	"basictree"
	"time"
)

func main() {
	// Just noting and printing the program start time 
	t1 := time.Now()
	fmt.Println("Time Start:", t1.Format(time.StampMilli))

	// For the files and folders
	folders := make([]string, 0)
	filestats := make([]fs.FileInfo, 0)

	// For processing arguments and options (if any)
	args := os.Args[1:]
	tempargs := make([]string, 0)
	options := make([]string, 0)

	if len(args) > 0 {

		// To check if the arguments have any file/directory names seperated by commas and if there are any options used
		for i, r := range args {
			if strings.Contains(r, ",") {  // if any argument has file/directory names seperated by comma
				tempargs = append(tempargs, strings.Split(r, ",")...)
			} else if r == "-full" {	// For the options
				options = append(options, args[i:]...)
				break
			} else { 
				tempargs = append(tempargs, r)
			}
		}
		// Rest all arguments back to arguments from temp slice for further processing
		args = tempargs 

		// To check if the file/directory names are valid and segregating files and directories
		for _, r := range args {
			r, _ = filepath.Abs(r)
			a, err := os.Stat(r)
			if err == nil && a.IsDir() {
				folders = append(folders, r)
			} else if err == nil {
				filestats = append(filestats, a)
			} else {
				fmt.Println(err)
			}
		}
	} 
	

	// Processing the options if there are any used
	var full bool = false
	if len(options) > 0 {
		for _, r := range options {
			if r == "-full" {
				full = true
			}
		}
	}

	// Processing the folders (if any)
	if len(folders) > 0 {
		folderNodes := make([]*FolderNode, 0)

		// Here we just walk through each folder and map it in our FolderNode struct.
		for _, r := range folders {
			node := walkDir(r)
			folderNodes = append(folderNodes, node)
		}

		// Here we calculate and state the total sizes and Print the tree as of right now
		for _, r := range folderNodes {
			// fmt.Println("d__", r.Name, "=>", prettyByteSize(r.TotalSize()))
			r.TotalSize()
			basictree.Tree(r.show(0, "%s [%s__f]", full), 4, 0, 1)
			fmt.Println()
		}

		// basictree.Tree(folderNodes[0].show(0, "%s [%s__f]"), 4, 0, 1)
		// fmt.Println(folderNodes[0].show(0, "%s [%s__f]"))
	}

	// Processing the files if any were mentioned specifically in the arguments.
	if len(filestats) > 0 {
		for _, r := range filestats {
			fmt.Println("f__", r.Name(), "=>", prettyByteSize(int(r.Size())))
		}
	}
	

	// Just the estimated time elapsed since the program started. 
	t2 := time.Now()
	fmt.Println("Time Elapsed:", t2.Sub(t1))
	
}



// Only for basictree.Tree() because it returns []basictree.Node
// 
// Takes in various options and returns []basictree.Node for basictree.Tree() to print the tree structure of the FolderNode given.
// 
// Options include, level for the start level of the node function to remember the depth
// 
// filefmt is used by fmt.Sprintf() to format how the file.Name and file.Size will be printed incase of files in the tree.
// 
// full options is used to see weather we print all the files and directories of the folder or just the one that occupy more than 20% of size of the parent directory 
func (f *FolderNode) show(level int, filefmt string, full bool) []basictree.Node {
	
	temp := make([]basictree.Node, 0)
	dcontent := fmt.Sprintf("%s [%s]", f.Name, prettyByteSize(f.Size))
	temp = append(temp, basictree.Node{Level: level, Content: dcontent})

	for _, file := range f.Files {
		if float64(file.Size) < float64(f.Size)*float64(0.2) && !full {
			continue
		}
		fcontent := fmt.Sprintf(/* "%s [%s__f]" */filefmt, file.Name, prettyByteSize(file.Size))
		temp = append(temp, basictree.Node{Level: level+1, Content: fcontent})
	}
	
	for _, child := range f.Children {
		if float64(child.Size) < float64(f.Size)*float64(0.2) && !full {
			continue
		}
		temp = append(temp, child.show(level+1, filefmt, full)...)
	}

	return temp
}


// function found on the stackover flow to take in number of bytes in form of int and return a string of the pretty KB, MB, GB... representation.
func prettyByteSize(b int) string {
	bf := float64(b)
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%6.4f %sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.4fYiB", bf)
}


// This is the main function of this program/package
// 
// Here it's a recursive function which takes in a directory name and maps its information to the FolderNode and returns it, 
// 
// If there are children, it maps them recursively to the children slice of the FolderNode which is a slice of []FolderNode.
func walkDir(dir string) *FolderNode {
	node := &FolderNode{Name: filepath.Base(dir)}

	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println(err)
	}

	for _, de := range files {
		f, err := de.Info()
		if err != nil {
			fmt.Println(err)
		}

		if f.IsDir() {
			child := walkDir(filepath.Join(dir, f.Name()))
			child.Parent = node
			node.Children = append(node.Children, child)	
		} else {
			file := File{Name: f.Name(), Size: int(f.Size())}
			node.Files = append(node.Files, file)
		}
	}

	return node
}

type FolderNode struct {
	Name 		string
	Size 		int

	Files 		[]File
	Children 	[]*FolderNode

	Parent 		*FolderNode
}

type File struct {
	Name 		string
	Size 		int
}


// After the Folder has been walked by the walkdir function, we call this function to calculate and incorporate the sizes of all the Folders and Files in the FolderNode
func (f *FolderNode) TotalSize() int {
	size := f.Size

	for _, file := range f.Files {
		size += file.Size 
	}

	for _, child := range f.Children {
		size += child.TotalSize()
	}

	f.Size = size

	return size
}

