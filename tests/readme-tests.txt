This directory contains integration tests for ictiobus. Each one is a shell
script that tests execution of a particular feature.

To run all tests, execute run-all.sh. Note that individual tests assume that
ictcc has been built already; run-all.sh will handle this for you.

To run an individual test, execute the exec.sh script in the test case's
directory.

To create a new test, add a new folder in this folder with an exec.sh file.

Once the test is written in the exec.sh file, run it and put its output in
expect.txt in that folder. Make sure to redirect both stderr and stdout to that
file (put 2>&1 at the end of the execution, after the file redirection).

All go code inside of an individual test's directory should be in a
sub-directory starting with a dot to prevent inclusion in runs of
`go test ./...`.
