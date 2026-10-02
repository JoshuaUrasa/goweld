package main

import (
	"os"
)






func New(name string) error{

		err := os.Mkdir(name,0750)

		if err != nil {
			if os.IsExist(err){

				return err
			} else {

				return err
			}

		} else {

				return  nil
}
}
