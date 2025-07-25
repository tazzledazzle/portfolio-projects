#!/usr/bin/env python3
"""
Test script to verify backend can start without Docker
"""
import sys
import os

# Add the app directory to Python path
sys.path.insert(0, os.path.join(os.path.dirname(__file__)))

def test_import():
    """Test that we can import the main app"""
    try:
        from app.main import app
        print("✓ Successfully imported FastAPI app")
        return True
    except Exception as e:
        print(f"✗ Failed to import app: {e}")
        return False

def test_health_endpoint():
    """Test that health endpoint works"""
    try:
        from app.main import health_check
        import asyncio
        
        result = asyncio.run(health_check())
        print(f"✓ Health check returned: {result}")
        return True
    except Exception as e:
        print(f"✗ Health check failed: {e}")
        return False

def main():
    """Run all tests"""
    print("Testing DevStack Manager Backend startup...")
    
    tests = [
        test_import,
        test_health_endpoint
    ]
    
    passed = 0
    for test in tests:
        if test():
            passed += 1
    
    print(f"\nResults: {passed}/{len(tests)} tests passed")
    
    if passed == len(tests):
        print("✓ Backend can start successfully!")
        return 0
    else:
        print("✗ Some tests failed")
        return 1

if __name__ == "__main__":
    sys.exit(main())