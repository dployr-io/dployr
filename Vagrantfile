# Vagrantfile - Simple dployr.io testing setup
# -*- mode: ruby -*-

# All distributions to test
DISTROS = {
  "ubuntu20" => "ubuntu/focal64",
  "ubuntu22" => "ubuntu/jammy64",
  "ubuntu24" => "ubuntu/noble64",
  "debian11" => "debian/bullseye64",
  "debian12" => "debian/bookworm64",
  "centos8"  => "centos/stream8",
  "rocky9"   => "rockylinux/9",
  "alma9"    => "almalinux/9",
  "opensuse" => "opensuse/Leap-15.5.x86_64",
  "arch"     => "archlinux/archlinux"
}

Vagrant.configure("2") do |config|
  config.vm.boot_timeout = 600  # 10 minutes
  config.vm.network "forwarded_port", guest: 7879, host: 7879
  config.vm.network "forwarded_port", guest: 8000, host: 8001
  config.vm.network "forwarded_port", guest: 6001, host: 6001
  config.vm.network "forwarded_port", guest: 6002, host: 6002
  config.vm.network "forwarded_port", guest: 22, host: 2222
  config.vm.network "forwarded_port", guest: 80, host: 8002
  config.vm.network "forwarded_port", guest: 443, host: 4430
  
  DISTROS.each_with_index do |(name, box), index|
    config.vm.define name do |vm|
      vm.vm.box = box
      vm.vm.hostname = "dployr-#{name}"
      
      # Network & ports
      vm.vm.network "private_network", ip: "192.168.56.#{10 + index}"
      vm.vm.network "forwarded_port", guest: 3000, host: 3000 + index
      
      # Resources
      vm.vm.provider "virtualbox" do |v|
        v.memory = 2048
        v.cpus = 2
        v.name = "dployr-#{name}"
        v.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
        v.customize ["modifyvm", :id, "--natdnsproxy1", "on"]
      end
      
      # System update
      vm.vm.provision "shell", inline: <<-SHELL
        echo "Updating #{name}..."
        if command -v apt >/dev/null; then
          apt update -qq && apt upgrade -y -qq
        elif command -v dnf >/dev/null; then
          dnf update -y -q
        fi
      SHELL
    end
  end
end