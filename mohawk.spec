%global provider        github
%global provider_tld    com
%global project         yaacov
%global repo            mohawk
# https://github.com/yaacov/mohawk
%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path     %{provider_prefix}
%global commit          8eefab3f90ab8c828e202a5a0fc20150ecae1ff2
%global shortcommit     %(c=%{commit}; echo ${c:0:7})

Name:           %{repo}
Version:        0.8.2
Release:        6%{?dist}
Summary:        Mock Hawkular server
License:        MIT
URL:            https://%{import_path}
Source0:        https://github.com/yaacov/mohawk/archive/%{version}.tar.gz

BuildRequires:  gcc

BuildRequires:  golang >= 1.2-7

# pull in golang libraries by explicit import path, inside the meta golang()
BuildRequires:  golang(github.com/mattn/go-sqlite3) >= 0-0.1

%description
Mock Hawkular server

%prep
%setup -q -n mohawk-%{version}

# many golang binaries are "vendoring" (bundling) sources, so remove them. Those dependencies need to be packaged independently.
rm -rf vendor

%build
# set up temporary build gopath, and put our directory there
mkdir -p ./_build/src/github.com/yaacov
ln -s $(pwd) ./_build/src/github.com/yaacov/mohawk

export GOPATH=$(pwd)/_build:%{gopath}
go build -o mohawk .

%install
install -d %{buildroot}%{_bindir}
install -p -m 0755 ./mohawk %{buildroot}%{_bindir}/mohawk

%files
%defattr(-,root,root,-)
%doc LICENSE README.md
%{_bindir}/mohawk
