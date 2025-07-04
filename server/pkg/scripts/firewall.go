package scripts

var FirewallSetupScript = []string{
	// Network operation with 3 retries
    `for i in {1..3}; do sudo apt-get update && break; sleep $((i*3)); done; [ $? -ne 0 ] && exit 1`,
    
    // Install with built-in recovery
    `sudo apt-get install -y --fix-broken ufw`,
    
    // Critical lock cleanup
    `sudo rm -f /run/ufw.lock 2>/dev/null`,
    
    // Reset to known state
    `sudo ufw --force reset`,
    
    // Configure essential rules
    `sudo ufw default deny incoming`,
    `sudo ufw default allow outgoing`,
    `sudo ufw allow 22/tcp`,   // SSH
    `sudo ufw allow 80/tcp`,   // HTTP
    `sudo ufw allow 443/tcp`,  // HTTPS
    
    // Enable firewall
    `sudo ufw --force enable`,
    
    // Final verification
    `sudo ufw status | grep -q "22/tcp" || exit 1`,
}