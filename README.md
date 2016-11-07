cdbfc - Go code to convert dBase files to CSV

# SYNOPSIS
#### For information about a DBase file
<code>**cdbfc** -i file.dbf</code>

#### For a single file to stdout
<code>**cdbfc** -f file.dbf > file.csv</code>

#### For a single file to the same file with .csv extension
<code>**cdbfc** -f file.dbf -o</code>

#### For multiple files in directory (to multiple files in the same directory)
<code>**cdbfc** -d directory</code>

#### For multiple files in directory concatenaed to stdout
<code>**cdbfc** -c directory</code>

#### For multiple files in directory concatenaed to stdout (with first column being the file name)
<code>**cdbfc** -c directory -a</code>

# DESCRIPTION
To build the package:

<code>go build cdbfc.go</code>
