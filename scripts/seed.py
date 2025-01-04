#!/bin/python
import os
import sys
import stat
import subprocess

KEIJI_CTL = "keiji-ctl"
URL = os.getenv("SITE_URL")

# the seed loop needs to create all of the:
#   - admin table links
#   - navbar png/redirect combos
#   - the menu link pairs

admin_table = {
        "new": {
            "blog post": "/admin/posts",
            "digital media": "/admin/upload"
            },
        "modify": {
            "blog post": "/admin/posts/all"
            },
        }
menu = {
        "//Administrator": "/admin/panel",
        "//Creative Writing": "/creative",
        "//Black Box": "/blog",
        "//Digital Art": "/digital"
        }


navbar_items = {
        "./assets/github.png": "https://github.com/AETH-erial",
        "./assets/git.png": "https://git.aetherial.dev/aeth",
        "./assets/twitter.png": "https://x.com/Aetherial___",
        "./assets/linkedin.png": "https://www.linkedin.com/in/russell-hrubesky-a62237221/",
        "./assets/soundcloud.png": "https://soundcloud.com/aeth-592553883"
            }

assets = [
        "./assets/menu.png",
        "./assets/github.png",
        "./assets/git.png",
        "./assets/twitter.png",
        "./assets/linkedin.png",
        "./assets/soundcloud.png"
        ]


# find the keiji-ctl command
def _find_keiji_ctl() -> str:
    split_path = os.environ["PATH"].split(":")
    for path in split_path:
        cmd_path = f"{path}/{KEIJI_CTL}"
        try:
            mode = os.stat(f"{path}/{KEIJI_CTL}").st_mode
        except FileNotFoundError:
            continue
        if stat.S_ISREG(mode):
            return cmd_path
    raise FileNotFoundError(f"the {KEIJI_CTL} binary could not be found in the system path.")


def main():
    """ setup script shit """
    path = _find_keiji_ctl()
    cookie = subprocess.run([path, "-cmd", "auth", "-address", URL], capture_output=True, text=True).stdout.strip("\n")
    print(cookie)
    for asset in assets:
        subprocess.run([path, "-address", URL, "-cookie", cookie, "-cmd", "asset", "-png", asset])
    for image, redirect in navbar_items.items():
        subprocess.run([path, "-address", URL, "-cookie", cookie, "-cmd", "nav", "-png", image, "-redirect", redirect])
    for text, redirect in menu.items():
        subprocess.run([path, "-address", URL, "-cookie", cookie, "-cmd", "menu", "-text", text, "-redirect", redirect])
    for category, pairings in admin_table.items():
        for text, redirect in pairings.items():
            subprocess.run([path, "-address", URL, "-cookie", cookie, "-cmd", "admin", "-text", text, "-redirect", redirect, "-col", category])

main()
