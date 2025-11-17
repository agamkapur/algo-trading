cmake_minimum_required(VERSION 3.15)
project(kucoin_connector)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(Boost REQUIRED COMPONENTS system json)
find_package(OpenSSL REQUIRED)

add_executable(kucoin_connector main.cpp)

target_link_libraries(kucoin_connector
    Boost::system
    Boost::json
    OpenSSL::SSL
    OpenSSL::Crypto
    pthread
)
