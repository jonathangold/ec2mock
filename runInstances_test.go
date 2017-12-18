// Mock EC2 Client
// Copyright (C) 2017+ Jonathan Gold and the project contributors
// Written by Jonathan Gold <info@jonathangold.ca> and the project
// contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package ec2mock

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestRunInstances(t *testing.T) {
	c := &MockEc2Client{}
	out, err := c.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String("ami-00000000"),
		InstanceType: aws.String("t2.micro"),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(5),
	})
	if err != nil {
		t.Error(err)
	}
	if len(out.Instances) != 5 {
		t.Error("incorrect number of instances in output")
	}
	if len(c.Reservations) != 1 {
		t.Error("incorrect number of reservations")
	}
	if len(c.Reservations[0].Instances) != 5 {
		t.Error("incorrect number of instances in reservation")
	}
}
