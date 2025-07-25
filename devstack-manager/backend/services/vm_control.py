import subprocess
import json
import shutil
from typing import List, Dict, Optional

class VMService:
    def __init__(self):
        self.vm_provider = self._detect_vm_provider()
    
    def _detect_vm_provider(self) -> str:
        """Detect available VM provider"""
        if shutil.which("VBoxManage"):
            return "virtualbox"
        elif shutil.which("virsh"):
            return "libvirt"
        elif shutil.which("vagrant"):
            return "vagrant"
        else:
            return "none"
    
    def list_vms(self) -> List[Dict]:
        """List all VMs based on available provider"""
        if self.vm_provider == "virtualbox":
            return self._list_virtualbox_vms()
        elif self.vm_provider == "libvirt":
            return self._list_libvirt_vms()
        elif self.vm_provider == "vagrant":
            return self._list_vagrant_vms()
        else:
            return []
    
    def start_vm(self, vm_name: str) -> Dict:
        """Start a VM"""
        if self.vm_provider == "virtualbox":
            return self._start_virtualbox_vm(vm_name)
        elif self.vm_provider == "libvirt":
            return self._start_libvirt_vm(vm_name)
        elif self.vm_provider == "vagrant":
            return self._start_vagrant_vm(vm_name)
        else:
            return {"status": "error", "message": "No VM provider available"}
    
    def stop_vm(self, vm_name: str) -> Dict:
        """Stop a VM"""
        if self.vm_provider == "virtualbox":
            return self._stop_virtualbox_vm(vm_name)
        elif self.vm_provider == "libvirt":
            return self._stop_libvirt_vm(vm_name)
        elif self.vm_provider == "vagrant":
            return self._stop_vagrant_vm(vm_name)
        else:
            return {"status": "error", "message": "No VM provider available"}
    
    def get_vm_status(self, vm_name: str) -> Dict:
        """Get VM status"""
        if self.vm_provider == "virtualbox":
            return self._get_virtualbox_vm_status(vm_name)
        elif self.vm_provider == "libvirt":
            return self._get_libvirt_vm_status(vm_name)
        elif self.vm_provider == "vagrant":
            return self._get_vagrant_vm_status(vm_name)
        else:
            return {"status": "error", "message": "No VM provider available"}
    
    # VirtualBox implementations
    def _list_virtualbox_vms(self) -> List[Dict]:
        """List VirtualBox VMs"""
        try:
            result = subprocess.run(
                ["VBoxManage", "list", "vms"],
                capture_output=True,
                text=True,
                check=True
            )
            
            vms = []
            for line in result.stdout.strip().split('\n'):
                if line:
                    # Parse: "VM Name" {uuid}
                    parts = line.split('" {')
                    if len(parts) == 2:
                        name = parts[0].strip('"')
                        uuid = parts[1].rstrip('}')
                        
                        # Get VM state
                        state_result = subprocess.run(
                            ["VBoxManage", "showvminfo", uuid, "--machinereadable"],
                            capture_output=True,
                            text=True
                        )
                        
                        state = "unknown"
                        if state_result.returncode == 0:
                            for state_line in state_result.stdout.split('\n'):
                                if state_line.startswith('VMState='):
                                    state = state_line.split('=')[1].strip('"')
                                    break
                        
                        vms.append({
                            "name": name,
                            "uuid": uuid,
                            "status": state,
                            "provider": "virtualbox"
                        })
            
            return vms
        except subprocess.CalledProcessError as e:
            return [{"error": f"Failed to list VirtualBox VMs: {e}"}]
        except Exception as e:
            return [{"error": f"Error listing VMs: {e}"}]
    
    def _start_virtualbox_vm(self, vm_name: str) -> Dict:
        """Start VirtualBox VM"""
        try:
            subprocess.run(
                ["VBoxManage", "startvm", vm_name, "--type", "headless"],
                check=True,
                capture_output=True
            )
            return {"status": "started", "name": vm_name, "provider": "virtualbox"}
        except subprocess.CalledProcessError as e:
            return {"status": "error", "message": f"Failed to start VM: {e}"}
    
    def _stop_virtualbox_vm(self, vm_name: str) -> Dict:
        """Stop VirtualBox VM"""
        try:
            subprocess.run(
                ["VBoxManage", "controlvm", vm_name, "poweroff"],
                check=True,
                capture_output=True
            )
            return {"status": "stopped", "name": vm_name, "provider": "virtualbox"}
        except subprocess.CalledProcessError as e:
            return {"status": "error", "message": f"Failed to stop VM: {e}"}
    
    def _get_virtualbox_vm_status(self, vm_name: str) -> Dict:
        """Get VirtualBox VM status"""
        try:
            result = subprocess.run(
                ["VBoxManage", "showvminfo", vm_name, "--machinereadable"],
                capture_output=True,
                text=True,
                check=True
            )
            
            status_info = {"name": vm_name, "provider": "virtualbox"}
            for line in result.stdout.split('\n'):
                if line.startswith('VMState='):
                    status_info["status"] = line.split('=')[1].strip('"')
                elif line.startswith('memory='):
                    status_info["memory"] = line.split('=')[1]
                elif line.startswith('cpus='):
                    status_info["cpus"] = line.split('=')[1]
            
            return status_info
        except subprocess.CalledProcessError:
            return {"status": "not_found", "name": vm_name}
    
    # Libvirt implementations
    def _list_libvirt_vms(self) -> List[Dict]:
        """List libvirt VMs"""
        try:
            result = subprocess.run(
                ["virsh", "list", "--all"],
                capture_output=True,
                text=True,
                check=True
            )
            
            vms = []
            lines = result.stdout.strip().split('\n')[2:]  # Skip header
            for line in lines:
                if line.strip():
                    parts = line.split()
                    if len(parts) >= 3:
                        vm_id = parts[0] if parts[0] != '-' else None
                        name = parts[1]
                        status = ' '.join(parts[2:])
                        
                        vms.append({
                            "name": name,
                            "id": vm_id,
                            "status": status,
                            "provider": "libvirt"
                        })
            
            return vms
        except subprocess.CalledProcessError as e:
            return [{"error": f"Failed to list libvirt VMs: {e}"}]
    
    def _start_libvirt_vm(self, vm_name: str) -> Dict:
        """Start libvirt VM"""
        try:
            subprocess.run(["virsh", "start", vm_name], check=True, capture_output=True)
            return {"status": "started", "name": vm_name, "provider": "libvirt"}
        except subprocess.CalledProcessError as e:
            return {"status": "error", "message": f"Failed to start VM: {e}"}
    
    def _stop_libvirt_vm(self, vm_name: str) -> Dict:
        """Stop libvirt VM"""
        try:
            subprocess.run(["virsh", "shutdown", vm_name], check=True, capture_output=True)
            return {"status": "stopped", "name": vm_name, "provider": "libvirt"}
        except subprocess.CalledProcessError as e:
            return {"status": "error", "message": f"Failed to stop VM: {e}"}
    
    def _get_libvirt_vm_status(self, vm_name: str) -> Dict:
        """Get libvirt VM status"""
        try:
            result = subprocess.run(
                ["virsh", "dominfo", vm_name],
                capture_output=True,
                text=True,
                check=True
            )
            
            status_info = {"name": vm_name, "provider": "libvirt"}
            for line in result.stdout.split('\n'):
                if line.startswith('State:'):
                    status_info["status"] = line.split(':', 1)[1].strip()
                elif line.startswith('Max memory:'):
                    status_info["max_memory"] = line.split(':', 1)[1].strip()
                elif line.startswith('Used memory:'):
                    status_info["used_memory"] = line.split(':', 1)[1].strip()
                elif line.startswith('CPU(s):'):
                    status_info["cpus"] = line.split(':', 1)[1].strip()
            
            return status_info
        except subprocess.CalledProcessError:
            return {"status": "not_found", "name": vm_name}
    
    # Vagrant implementations
    def _list_vagrant_vms(self) -> List[Dict]:
        """List Vagrant VMs"""
        try:
            result = subprocess.run(
                ["vagrant", "global-status"],
                capture_output=True,
                text=True,
                check=True
            )
            
            vms = []
            lines = result.stdout.strip().split('\n')
            in_vm_section = False
            
            for line in lines:
                if line.startswith('id'):
                    in_vm_section = True
                    continue
                elif line.startswith('---'):
                    in_vm_section = False
                    continue
                
                if in_vm_section and line.strip():
                    parts = line.split()
                    if len(parts) >= 4:
                        vms.append({
                            "name": parts[1],
                            "id": parts[0],
                            "status": parts[3],
                            "provider": "vagrant",
                            "directory": parts[4] if len(parts) > 4 else ""
                        })
            
            return vms
        except subprocess.CalledProcessError as e:
            return [{"error": f"Failed to list Vagrant VMs: {e}"}]
    
    def _start_vagrant_vm(self, vm_name: str) -> Dict:
        """Start Vagrant VM"""
        try:
            subprocess.run(["vagrant", "up", vm_name], check=True, capture_output=True)
            return {"status": "started", "name": vm_name, "provider": "vagrant"}
        except subprocess.CalledProcessError as e:
            return {"status": "error", "message": f"Failed to start VM: {e}"}
    
    def _stop_vagrant_vm(self, vm_name: str) -> Dict:
        """Stop Vagrant VM"""
        try:
            subprocess.run(["vagrant", "halt", vm_name], check=True, capture_output=True)
            return {"status": "stopped", "name": vm_name, "provider": "vagrant"}
        except subprocess.CalledProcessError as e:
            return {"status": "error", "message": f"Failed to stop VM: {e}"}
    
    def _get_vagrant_vm_status(self, vm_name: str) -> Dict:
        """Get Vagrant VM status"""
        try:
            result = subprocess.run(
                ["vagrant", "status", vm_name],
                capture_output=True,
                text=True,
                check=True
            )
            
            # Parse vagrant status output
            for line in result.stdout.split('\n'):
                if vm_name in line:
                    parts = line.split()
                    if len(parts) >= 2:
                        status = ' '.join(parts[1:])
                        return {
                            "name": vm_name,
                            "status": status,
                            "provider": "vagrant"
                        }
            
            return {"status": "not_found", "name": vm_name}
        except subprocess.CalledProcessError:
            return {"status": "not_found", "name": vm_name}