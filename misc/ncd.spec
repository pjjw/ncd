#
# rpm spec file for ncd
#

%global debug_package %{nil}
%global __spec_install_post /usr/lib/rpm/check-rpaths\
/usr/lib/rpm/check-buildroot  \


Summary: ncd
Name: ncd
Version: 0.0.1
Release: 1%{?redhatvers:.%{redhatvers}}
License: BSD
Group: system
Packager: peter woodman <pjjw@google.com>
Source0: ncd
Source1: ncd.init
Source2: ncd.options

BuildRoot: %{_tmppath}/%{name}-%{version}-root

%description
pair of utilities to inject checks into nagios' checkresult spool

%changelog
* Mon Jul 13 2011 Peter Woodman <pjjw@google.com> 0.0.1-1
- initial gross packaging

%prep

%build

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT
install -D -m 755 %{_sourcedir}/ncd %{buildroot}/usr/bin/ncd
install -D -m 755 %{_sourcedir}/ncd.init %{buildroot}/etc/init.d/ncd
install -D -m 644 %{_sourcedir}/ncd.options %{buildroot}/etc/sysconfig/ncd


%clean


%files
%defattr(-,root,root)
%{_sbindir}/ncd
%{_bindir}/send_check
%{_sysconfdir}/init.d/ncd
%config(noreplace) %{_sysconfdir}/sysconfig/ncd

