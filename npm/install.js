#!/usr/bin/env node

const { existsSync, chmodSync, createWriteStream, unlinkSync } = require('fs');
const { get } = require('https');
const { join } = require('path');
const { pipeline } = require('stream');
const { promisify } = require('util');
const tar = require('tar');

const streamPipeline = promisify(pipeline);

const BINARY_NAME = 'misaki-banner';
const REPO = 'qraqras/misaki-banner';

function getPlatform() {
    const platform = process.platform;
    const arch = process.arch;

    const platformMap = {
        darwin: 'darwin',
        linux: 'linux',
        win32: 'windows'
    };

    const archMap = {
        x64: 'x86_64',
        arm64: 'arm64'
    };

    const mappedPlatform = platformMap[platform];
    const mappedArch = archMap[arch];

    if (!mappedPlatform || !mappedArch) {
        throw new Error(`Unsupported platform: ${platform} ${arch}`);
    }

    return { platform: mappedPlatform, arch: mappedArch };
}

function getVersion() {
    const pkg = require('./package.json');
    return pkg.version;
}

function getBinaryName() {
    const { platform } = getPlatform();
    return platform === 'windows' ? `${BINARY_NAME}.exe` : BINARY_NAME;
}

function getDownloadURL() {
    const version = getVersion();
    const { platform, arch } = getPlatform();

    const ext = platform === 'windows' ? 'zip' : 'tar.gz';
    const fileName = `misaki-banner_${version}_${platform}_${arch}.${ext}`;

    return `https://github.com/${REPO}/releases/download/v${version}/${fileName}`;
}

async function download(url, destPath) {
    return new Promise((resolve, reject) => {
        get(url, (response) => {
            if (response.statusCode === 302 || response.statusCode === 301) {
                // Follow redirect
                download(response.headers.location, destPath).then(resolve).catch(reject);
                return;
            }

            if (response.statusCode !== 200) {
                reject(new Error(`Download failed: ${response.statusCode} ${response.statusMessage}`));
                return;
            }

            const fileStream = createWriteStream(destPath);
            streamPipeline(response, fileStream)
                .then(resolve)
                .catch(reject);
        }).on('error', reject);
    });
}

async function extractTarGz(archivePath, destDir) {
    await tar.x({
        file: archivePath,
        cwd: destDir
    });
}

async function install() {
    try {
        const binDir = __dirname;
        const binaryName = getBinaryName();
        const binaryPath = join(binDir, binaryName);

        // Skip if already exists
        if (existsSync(binaryPath)) {
            console.log(`Binary already exists: ${binaryPath}`);
            return;
        }

        console.log('Downloading misaki-banner binary...');
        const url = getDownloadURL();
        console.log(`URL: ${url}`);

        const { platform } = getPlatform();
        const archivePath = join(binDir, platform === 'windows' ? 'archive.zip' : 'archive.tar.gz');

        await download(url, archivePath);
        console.log('Download complete. Extracting...');

        if (platform === 'windows') {
            // TODO: Implement zip extraction for Windows
            const AdmZip = require('adm-zip');
            const zip = new AdmZip(archivePath);
            zip.extractAllTo(binDir, true);
        } else {
            await extractTarGz(archivePath, binDir);
        }

        // Make executable
        if (platform !== 'windows') {
            chmodSync(binaryPath, 0o755);
        }

        // Clean up archive
        unlinkSync(archivePath);

        console.log('Installation complete!');
    } catch (error) {
        console.error('Installation failed:', error.message);
        console.error('You can download the binary manually from:');
        console.error(`https://github.com/${REPO}/releases`);
        process.exit(1);
    }
}

install();
