"use client";

import SmallLoader from "@/components/loader/SmallLoader";
import { useToast } from "@/hooks/use-toast";
import axios from "axios";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, FormEvent } from "react";
import ProboLogo from '../../asssets/images/probo-logo.png'
import Image from "next/image";

function Signin() {
    const router = useRouter();
    const { toast } = useToast();
    const [mail, setMail] = useState("");
    const [password, setPassword] = useState("");
    const [loading, setLoading] = useState(false);

    const regex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault()
        try {
            setLoading(true);
            if (!mail || !regex.test(mail) || !password || password.length < 8) {
                toast({
                    variant: "destructive",
                    title:
                        "Email should be proper or password must be at least 8 characters",
                });
                return false;
            }
            const SERVER_BASE_URL = process.env.NEXT_PUBLIC_SERVER_BASE_URL;
            const userRes = await axios.post(`${SERVER_BASE_URL}/user/login`, {
                email: mail,
                password: password,
            });

            const user = userRes.data;
            localStorage.setItem("token", user.token);

            if (!user.profile.onboarding.Bool) {
                router.push("/onboarding?step=1");
            } else {
                router.push("/dashboard");
            }
        } catch (error) {
            console.log(error)
            toast({
                variant: "destructive",
                title:
                    "Invalid email or password",
            });
            return false;
        } finally {
            setLoading(false);
        }
    };

    return (
        <section className="bg-gray-50 dark:bg-gray-900">
            <div className="flex flex-col items-center justify-center px-6 py-8 mx-auto md:h-screen lg:py-0">
                <Link href="/" className="flex items-center mb-6 text-2xl font-semibold text-gray-900 dark:text-white">
                    <Image
                        className="w-[15rem] h-[2.5rem] mr-2"
                        src={ProboLogo}
                        alt="logo"
                        width={240}
                        height={40}
                    />
                </Link>
                <div className="w-full bg-white rounded-lg shadow dark:border md:mt-0 sm:max-w-md xl:p-0 dark:bg-gray-800 dark:border-gray-700">
                    <div className="p-6 space-y-4 md:space-y-6 sm:p-8">
                        <h1 className="text-xl font-bold leading-tight tracking-tight text-gray-900 md:text-2xl dark:text-white text-center">
                            Sign in to your account
                        </h1>
                        <form className="space-y-4 md:space-y-6" onSubmit={handleSubmit}>
                            <div>
                                <label
                                    htmlFor="email"
                                    className="block mb-2 text-sm font-medium text-gray-900 dark:text-white"
                                >
                                    Your email
                                </label>
                                <input
                                    type="email"
                                    name="email"
                                    id="email"
                                    value={mail}
                                    onChange={(e) => setMail(e.target.value)}
                                    className="bg-gray-50 border border-gray-300 text-gray-900 rounded-lg focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                                    placeholder="name@company.com"
                                    required
                                />
                            </div>
                            <div>
                                <label
                                    htmlFor="password"
                                    className="block mb-2 text-sm font-medium text-gray-900 dark:text-white"
                                >
                                    Password
                                </label>
                                <input
                                    type="password"
                                    name="password"
                                    id="password"
                                    placeholder="••••••••"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    className="bg-gray-50 border border-gray-300 text-gray-900 rounded-lg focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                                    required
                                />
                            </div>
                            <button
                                type="submit"
                                className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5 me-2 mb-2 dark:bg-blue-600 dark:hover:bg-blue-700 focus:outline-none dark:focus:ring-blue-800 w-full flex justify-center">
                                {loading ? <SmallLoader size={20} /> : "Sign in"}
                            </button>
                            <p className="text-sm font-light text-gray-500 dark:text-gray-400">
                                Don’t have an account yet?{' '}
                                <Link href="/signup" className="text-blue-500">
                                    Sign Up
                                </Link>
                            </p>
                        </form>
                    </div>
                </div>
            </div>
        </section>
    );
}
export default Signin;