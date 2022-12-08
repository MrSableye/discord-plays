#include "gb_headlesswrapper.hxx"
#include "httplib.h"
#include <iostream>
#include <string>
#include <mutex>

Command serialize(std::string command) {
    if (command == "start")
        return Command::Start;
    else if (command == "select")
        return Command::Select;
    else if (command == "a")
        return Command::A;
    else if (command == "b")
        return Command::B;
    else if (command == "d")
        return Command::Down;
    else if (command == "u")
        return Command::Up;
    else if (command == "l")
        return Command::Left;
    else if (command == "r")
        return Command::Right;
    else if (command == "exit")
        return Command::Exit;
    else if (command == "screen")
        return Command::ScreenshotPNG;
    else if (command == "save")
        return Command::Save;
    else if (command == "load")
        return Command::Load;
    else if (command == "read")
        return Command::ReadSingle;
    else if (command == "string")
        return Command::ReadString;
    else if (command == "spam")
        return Command::Spam;
    return Command::Error;
}

int main(int argc, char** argv) {
    if (argc != 2) {
        std::cerr << "Invalid number of arguments" << std::endl;
        return 1;
    }
    std::string path = argv[1];
    Gameboy gb(path);
    gb.ExecuteCommand(Command::Screenshot);
    gb.ExecuteCommand(Command::Start);
    gb.ExecuteCommand(Command::Second);
    gb.ExecuteCommand(Command::Start);
    gb.ExecuteCommand(Command::A);
    gb.ExecuteCommand(Command::Second);
    gb.ExecuteCommand(Command::A);
    gb.ExecuteCommand(Command::A);
    gb.ExecuteCommand(Command::A);
    
    gb.SetMemory(0xD199, 0); // Set text speed to instant
    httplib::Server svr;
    std::mutex the_mutex;
    svr.Get("/req", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.SetValue("");
        res.set_content("ack", "text/plain");
        auto str = req.get_param_value("action");
        bool expect_ret = (req.get_param_value_count("val") != 0) || str == "screen";
        if (expect_ret)
            gb.SetValue(req.get_param_value("val"));
        auto com = serialize(str);
        gb.ExecuteCommand(com);
        if (expect_ret)
            res.set_content(gb.GetRes().c_str(), "text/plain");
    });
    svr.Get("/party", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::GetParty);
        res.set_content(gb.GetRes(), "text/plain");
    });
    svr.Get("/trainer", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::GetTrainer);
        res.set_content(gb.GetRes(), "text/plain");
    });
    svr.Get("/balls", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::GetBalls);
        res.set_content(gb.GetRes(), "text/plain");
    });
    svr.Get("/map", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::GetMap);
        res.set_content(gb.GetRes(), "text/plain");
    });
    svr.Get("/save", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::StartSave);
        res.set_content("started save", "text/plain");
    });
    svr.Get("/gif", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::GetGif);
        res.set_content("saved gif", "text/plain");
    });
    svr.listen("localhost", 1234);
}