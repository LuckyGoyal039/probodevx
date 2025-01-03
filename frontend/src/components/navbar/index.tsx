'use client'
import Image from "next/image";
import { Menu } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";
// import { useRecoilValue } from "recoil";
// import { isAuthenticate } from "@/store/atoms/Auth";

import userAvatar from '../../asssets/images/probo-logo.png';
import proboLogo from '../../asssets/images/probo-logo.png'

interface NavLink {
    href: string,
    lable: string
}

const navOptions: NavLink[] = [
    { href: "/events", lable: "Trading" },
    { href: "/team-11", lable: "Team 11" },
    { href: "/read", lable: "Read" },
    { href: "/cares", lable: "Cares" },
]
export default function Navbar() {
    const [isFixed, setIsFixed] = useState(false);
    // const isAuth = useRecoilValue(false);
    const isAuth = false;
    useEffect(() => {
        const handleScroll = () => {
            const scrollThreshold = 10;
            if (window.scrollY > scrollThreshold) {
                setIsFixed(true);
            } else {
                setIsFixed(false);
            }
        };
        window.addEventListener('scroll', handleScroll);

        return () => {
            window.removeEventListener('scroll', handleScroll);
        };
    }, []);

    return (
        <div className={`w-full transition-all duration-500 z-10 ${isFixed ? 'fixed top-2 px-4' : ''}`}>
            <div
                className={`flex justify-between items-center bg-black/60 backdrop-blur-md border border-gray-800 z-10 p-4 md:px-10 
                    ${isFixed ? 'rounded-lg py-2 px-3' : ''}`}
            >
                <div className="flex items-center gap-2">
                    <div className="w-30 h-8  overflow-hidden rounded-lg bg-white flex items-center px-10 py-2">
                        <Link href={'/'}>
                            <Image
                                src={proboLogo}
                                alt="Main Logo"
                                objectFit="cover"
                                width={70}
                                height={30}
                            />
                        </Link>
                    </div>
                </div>
                <div className="flex items-center gap-6 text-sm text-white">
                    {
                        navOptions.map((link) => {
                            return (<div className="hidden md:block">
                                <Link href={link.href}>{link.lable}</Link>
                            </div>)
                        })
                    }
                    {isAuth ? (
                        <Image
                            className="w-10 h-10 rounded-full"
                            src={userAvatar}
                            alt="Rounded avatar"
                        />
                    ) : (
                        <>
                            <div className="bg-gray-800 p-2 rounded-lg flex gap-1 items-center">
                                <Link href={'/signin'}>Log in</Link>
                                <span className="bg-gray-700 px-2 py-1 text-xs h-5 rounded">L</span>
                            </div>
                            <div className="bg-gray-300 p-2 rounded-lg text-black">
                                <Link href={'/signup'}>Sign up</Link>
                            </div>
                        </>
                    )}
                    <div className="md:hidden">
                        <Menu />
                    </div>
                </div>
            </div>
        </div>
    );
}
