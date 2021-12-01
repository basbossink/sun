#! /usr/bin/env sh
current_wd=$(pwd)
build_dir="build"
trap 'cd ${current_wd}' EXIT
version=$(git describe --tags --exact-match)
hash=$(git rev-parse HEAD)
ldflags="-s -w -X main.Version=${version} -X main.CommitHash=${hash}"
readme="README.adoc"
license="LICENSE"
prog="sun"
for line in $(go tool dist list); do
  GOOS="${line%/*}"
  GOARCH="${line#*/}"
  extension=""
  if [ "${GOOS}" = "windows" ]; then
    extension=".exe"
  fi
  if [ "${GOOS}" = "ios" ] || [ "${GOOS}" = "android" ] ; then
    continue
  fi
  subdir="${prog}_${GOOS}_${GOARCH}"
  output_file_name="${build_dir}/${subdir}/${prog}${extension}"
  env "GOOS=${GOOS}" "GOARCH=${GOARCH}" "CGO_ENABLED=0" go build -ldflags "${ldflags}" -o "${output_file_name}"
  cd "${build_dir}" || exit 
  ln -s "../../${readme}" "${subdir}/${readme}"
  ln -s "../../${license}" "${subdir}/${license}"
  zip -9 -r "${subdir}.zip" "${subdir}"
  cd "${current_wd}" || exit 
done

