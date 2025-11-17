cmake_minimum_required(VERSION 3.15)
project(crypto_com_connector)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(Boost REQUIRED COMPONENTS system)
find_package(OpenSSL REQUIRED)

add_executable(crypto_com_connector main.cpp)

target_link_libraries(crypto_com_connector
    Boost::system
    OpenSSL::SSL
    OpenSSL::Crypto
    pthread
)
