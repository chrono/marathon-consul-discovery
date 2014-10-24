# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

$script = <<SCRIPT
  sudo apt-get update
  sudo apt-get -y install golang git

  sudo mkdir -p /opt/gopath
  cat << EOF >/tmp/gopath.sh
export GOPATH="/opt/gopath"
export PATH="/opt/gopath/bin:\$PATH"
EOF
  sudo mv /tmp/gopath.sh /etc/profile.d/gopath.sh
  sudo chmod 0755 /etc/profile.d/gopath.sh
  sudo chown -R vagrant:vagrant /opt/gopath
SCRIPT

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "ubuntu/trusty64"

  config.vm.provision "shell", inline: $script

  config.vm.synced_folder ".", "/opt/gopath/src/github.com/chrono/marathon-consul-discovery"
end
