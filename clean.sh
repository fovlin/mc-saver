sudo rm -r .git
git init
git branch -m main
git remote add origin git@github.com:fovlin/mc-saver.git
git add *
git commit -am "Clean repo"
git push --set-upstream origin main -f
git push -f