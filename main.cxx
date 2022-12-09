#include "gb_headlesswrapper.hxx"
#include "httplib.h"
#include <iostream>
#include <string>
#include <mutex>

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
    svr.Get("/move_left", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::Left);
        res.set_content("ok", "text/plain");
    });
    svr.Get("/move_right", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::Right);
        res.set_content("ok", "text/plain");
    });
    svr.Get("/move_up", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::Up);
        res.set_content("ok", "text/plain");
    });
    svr.Get("/move_down", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::Down);
        res.set_content("ok", "text/plain");
    });
    svr.Get("/action_a", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::A);
        res.set_content("ok", "text/plain");
    });
    svr.Get("/action_b", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::B);
        res.set_content("ok", "text/plain");
    });
    svr.Get("/action_start", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::Start);
        res.set_content("ok", "text/plain");
    });
    svr.Get("/action_select", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::Select);
        res.set_content("ok", "text/plain");
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
        res.set_content(gb.GetRes(), "text/plain");
    });
    svr.Get("/load", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::Load);
        res.set_content(gb.GetRes(), "text/plain");
    });
    svr.Get("/gif", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::GetGif);
        res.set_content("saved gif", "text/plain");
    });
    svr.Get("/ping", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        res.set_content("pong", "text/plain");
    });
    svr.Get("/screen", [&gb, &the_mutex](const httplib::Request & req, httplib::Response &res) {
        std::lock_guard lg(the_mutex);
        gb.ExecuteCommand(Command::ScreenshotPNG);
        res.set_content(gb.GetRes(), "text/plain");
    });
    svr.listen("localhost", 1234);
}