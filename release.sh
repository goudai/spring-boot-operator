git fetch --tags

#echo -e "所有tag列表"
#git tag -l -n


#echo -e "${tagList}"
#获取最新版本tag
LatestTag=$(git describe --tags `git rev-list --tags --max-count=1`)

echo -e "最新版本tag......"
echo -e "$LatestTag"

echo -e "请输入要新增的版本号...... 如 v1.0.1"
#输入tag名称
read tagName

git tag -a ${tagName} -m ${tagName}
#推到分支上
git push origin ${tagName}