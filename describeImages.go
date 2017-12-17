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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func (c *MockEc2Client) DescribeImages(input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	time.Sleep(DescribeDelay * time.Second)
	return &ec2.DescribeImagesOutput{
		Images: []*ec2.Image{
			{
				ImageId: aws.String("ami-00000000"),
			},
		},
	}, nil
}
