#!/usr/bin/perl

use strict;
use warnings;
use File::Copy qw(copy);

local $ENV{PATH} = "$ENV{PATH};C:/Program Files/7-Zip";

my $version = '1.0';
my @OS = (
    'windows',
    'linux',
    'darwin',
    'freebsd',
);

my @ARCH = (
    '386',
    'amd64',
    'arm',
);

my @static = (
    'README.md',
    'LICENSE.txt',
);

mkdir 'tmp';
mkdir 'builds';

foreach my $s (@static) {
    copy($s, "tmp/$s");
}

foreach my $o (@OS) {
    my $ext = '';
    $ext = '.exe' if ($o eq 'windows');

    $ENV{'GOOS'} = $o;

    foreach my $a (@ARCH) {
        next if ($a eq 'arm' && $o ne 'linux');

        $ENV{'GOARCH'} = $a;

        print "Building ${o}/${a}\n";
        my $bin = "thing_v${version}_${o}_${a}";
        `go build -o tmp/${bin}${ext}`;

        if ($o eq 'windows') {
            `7z a builds/${bin}.zip ./tmp/*`;
        } else {
            `7z a builds/${bin}.tar ./tmp/*`;
            `7z a builds/${bin}.tar.gz ./builds/${bin}.tar`;
            unlink "./builds/${bin}.tar";
        }

        unlink "./tmp/${bin}${ext}";
    }
}

unlink glob "./tmp/*";
rmdir "./tmp";
