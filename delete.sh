for (( i = 2; i < 8; i++ )); do
V='v0.0.'${i}
git tag -d $V
git push origin :refs/tags/$V
done

