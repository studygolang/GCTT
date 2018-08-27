#!/bin/sh

git remote add upstream https://github.com/studygolang/GCTT
git fetch upstream
git checkout master
git rebase upstream/master
git push -f origin master
