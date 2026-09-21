"""Own-child process controls for E2E, including real Windows console events.

POSIX: SIGINT cancels CLI commands, SIGTERM drains the server.
Windows: give each long-lived child its own process group and send CTRL_BREAK.
Never broadcast console events to group 0 and never kill by executable name.
"""
from __future__ import annotations

from contextlib import contextmanager
import os
import signal
import subprocess


@contextmanager
def console_for_children():
    """Hosted Windows runners may launch Python without an attached console."""
    allocated = False
    kernel = None
    if os.name == "nt":
        import ctypes
        kernel = ctypes.WinDLL("kernel32", use_last_error=True)
        kernel.GetConsoleCP.restype = ctypes.c_uint
        kernel.AllocConsole.restype = ctypes.c_int
        kernel.FreeConsole.restype = ctypes.c_int
        if kernel.GetConsoleCP() == 0:
            if not kernel.AllocConsole():
                raise ctypes.WinError(ctypes.get_last_error())
            allocated = True
    try:
        yield
    finally:
        if allocated and kernel is not None:
            kernel.FreeConsole()


def start(argv, **kwargs) -> subprocess.Popen:
    if os.name == "nt":
        kwargs["creationflags"] = subprocess.CREATE_NEW_PROCESS_GROUP
    else:
        kwargs["start_new_session"] = True
    return subprocess.Popen(argv, **kwargs)


def interrupt(process: subprocess.Popen, *, server=False) -> None:
    if process.poll() is not None:
        raise AssertionError("Child exited before its cancellation/shutdown test")
    if os.name == "nt":
        # Requires console_for_children plus CREATE_NEW_PROCESS_GROUP.
        process.send_signal(signal.CTRL_BREAK_EVENT)
    else:
        process.send_signal(signal.SIGTERM if server else signal.SIGINT)


def force_cleanup(process: subprocess.Popen, timeout=5) -> None:
    """Last-resort cleanup is never counted as a successful graceful-stop test."""
    if process.poll() is None:
        process.terminate()
    try:
        process.wait(timeout=timeout)
    except subprocess.TimeoutExpired:
        process.kill()
        process.wait(timeout=timeout)


def graceful_stop(process: subprocess.Popen, timeout=15) -> int:
    interrupt(process, server=True)
    try:
        return process.wait(timeout=timeout)
    except subprocess.TimeoutExpired as exc:
        force_cleanup(process)
        raise AssertionError("Server did not exit within its graceful-shutdown deadline") from exc
