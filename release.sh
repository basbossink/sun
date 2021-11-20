#! /usr/bin/env sh
version=`git describe --tags --exact-match`
hash=`git rev-parse HEAD`
ldflags="-s -w -X main.Version=${version} -X main.CommitHash=${hash}"
build_dir="build"
readme="README.adoc"
license="LICENSE"
prog="sun"
for line in `go tool dist list`; do
  GOOS=$(cut -d'/' -f1 <<< $line)
  GOARCH=$(cut -d'/' -f2 <<< $line)
  extension=""
  if [ $GOOS == "windows" ]; then
    extension=".exe"
  fi
  if [ $GOOS == "ios" -o $GOOS == "android" ]; then
    continue
  fi
  subdir="${prog}_${GOOS}_${GOARCH}"
  output_file_name="${build_dir}/${subdir}/${prog}${extension}"
  env GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build -ldflags "${ldflags}" -o "${output_file_name}"
  pushd ${build_dir}
  ln -s "../../${readme}" "${subdir}/${readme}"
  ln -s "../../${license}" "${subdir}/${license}"
  zip -9 -r "${subdir}.zip" ${subdir}
  popd
done
